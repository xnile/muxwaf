local require       = require
local sites         = require("sites")
local log           = require("log")
local ngx_balancer  = require("ngx.balancer")
local roundrobin    = require("resty.balancer.roundrobin")
local lrucache      = require("resty.lrucache")
local cjson         = require("cjson.safe")
local net           = require("utils.net")
local dns           = require("utils.net.dns")
local ngx           = ngx
local ngx_exit      = ngx.exit
local ipairs        = ipairs
local tab_new       = table.new
local table_clone    = require("table.clone")

local HTTP_BAD_GATEWAY          = ngx.HTTP_BAD_GATEWAY
local HTTP_GATEWAY_TIMEOUT      = ngx.HTTP_GATEWAY_TIMEOUT

local _M = { 
  _VERSION = 0.1
}

local balancers = {}
local site_origins   = {}
local upstream_servers = {}


function _M.add_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to add origin, '", host, "' origin already exists")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
end


function _M.del_origins(_, host)
    if not origins[host] then
        log.warn("failed to delete '", host, "' origins, the site does not exist")
    end

    site_origins[host] = nil
end


function _M.update_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to update origins, '", host, "' origin already exists")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
end


function _M.full_sync_origins(_, items)
    local new_origins = {}
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins
        new_origins[host] = table_clone(origins)
    end
    site_origins = new_origins
end




-- local upstreams = {}

-- local CACHE_TTL  = 60
-- local CACHE_SIZE = 100
-- local DEFAULT_SERVER_WEIGHT = 100

-- local cache, err = lrucache.new(CACHE_SIZE * 2)
-- if not cache then
--     error("failed to create the cache: " .. (err or "unknown"), 2)
-- end


local function parse_orgins(origins)
    local servers = {
        http  = tab_new(0, #origins),
        https = tab_new(0, #origins)
    }

    for _, origin in ipairs(origins) do
        local addr_http  = origin.host .. ":" .. origin.http_port
        local addr_https = origin.host .. ":" .. origin.https_port
        local weight = origin.weight > 0 and origin.weight or DEFAULT_SERVER_WEIGHT
        servers.http[addr_http] = weight
        servers.https[addr_https] = weight
    end
    return servers
end


local function sync_lookup_origins()
    for host, _origins in pairs(site_origins) do

        local servers = {}
        for _, o in ipairs(_origins) do
            local ip_or_domain, port, weight  = o.host, o.http_port, o.weight

            -- domain need to dns lookup
            if not net.is_valid_ip(ip_or_domain) then
                local addresses, err = dns.lookup(ip_or_domain)
                if not addresses then
                    return nil, err
                end

                for _, ip in ipairs(addresses) do
                    local id = ip .. ":" .. port

                    -- TODO: Duplicate case
                    servers[id] = weight
                end
            else
                -- ip case
                local id = ip_or_domain .. ":" .. port
                servers[id] = weight
            end
        end


        upstream_servers[host] = servers
    end
end


-- local function get_upstream_servers(host)
--     local dns           = require("utils.net.dns")
--     local origins = site_origins[host]
--     if not origins then
--         return nil, "The origin server of the site'" .. host .. "'is empty"
--     end

--     local servers = {}
--     for _, o in ipairs(origins) do
--         local ip_or_domain, port, weight  = o.host, o.http_port, o.weight

--         -- domain need to dns lookup
--         if not net.is_valid_ip(ip_or_domain) then
--             local addresses, err = dns.lookup(ip_or_domain)
--             if not addresses then
--                 return nil, err
--             end

--             for _, ip in ipairs(addresses) do
--                 local id = ip .. ":" .. port

--                 -- TODO: Duplicate case
--                 servers[id] = weight
--             end
--         end

--         -- ip case
--         local id = ip_or_domain .. ":" .. port
--         servers[id] = weight
--     end

--     return servers, nil
-- end


local function get_balancer(host)
    if balancers[host] then
        return balancers[host]
    end


    local servers= upstream_servers[host]
    if not servers then
        return nil
    end

    balancers[host] = roundrobin:new(servers)

    return balancers[host]
end


-- local function pick_upstream_peer(host, upstream_protocol)
--     -- local origins = sites.get_site_origins(host)
--     -- if not origins then
--     --     return nil
--     -- end

--     local servers = cache:get(host)
--     if not servers then
--         servers = parse_orgins(origins)
--         cache:set(host, servers, CACHE_TTL)
--     end

--     local candidate = servers[upstream_protocol]
--     if not candidate then
--         return nil
--     end

--     local rr = roundrobin:new(candidate)
--     local peer = rr:find()

--     if not peer then
--         return nil
--     end

--     return peer
-- end


-- local function get_upstream(host, protocol)
--     return upstreams[host][protocol]
-- end

function _M.init_worker()
    local ok, err = ngx.timer.at(0, sync_lookup_origins)
    if not ok then
        ngx.log(ngx.ERR, "failed to create timer: ", err)
    end
end


function _M.balance(ctx)
    local host = ctx.host

    if not host then
        log.error("failed to get host")
        return ngx_exit(HTTP_GATEWAY_TIMEOUT)
    end

    ngx_balancer.set_more_tries(1)

    local balancer, err = get_balancer(host)
    if not balancer then
        log.error(err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end

    local peer = balancer:find()

    -- local peer = pick_upstream_peer(ctx.host, ctx.upstream_scheme)
    -- if not peer then
    --     return ngx_exit(HTTP_BAD_GATEWAY)
    -- end

    -- ok, err = balancer.set_current_peer(host, port)
    -- Domain names in host do not make sense. 
    local ok, err = ngx_balancer.set_current_peer(peer)
    if not ok then
        log.error("failed to setting current upstream peer \"", peer, "\", ", err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end
    log.debug("current upstream peer \"", peer, "\"")
end


return _M