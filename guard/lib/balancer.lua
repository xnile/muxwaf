local require       = require
local log           = require("log")
local ngx_balancer  = require("ngx.balancer")
local roundrobin    = require("resty.balancer.roundrobin")
-- local lrucache      = require("resty.lrucache")
local cjson         = require("cjson.safe")
local net           = require("utils.net")
local dns           = require("utils.net.dns")
local table_clone   = require("table.clone")
local ngx           = ngx
local ngx_exit      = ngx.exit
local ipairs        = ipairs
local tab_new       = table.new

local _M = { 
  _VERSION = 0.1
}

local HTTP_BAD_GATEWAY          = ngx.HTTP_BAD_GATEWAY
local HTTP_GATEWAY_TIMEOUT      = ngx.HTTP_GATEWAY_TIMEOUT
local SYNC_RESOLVE_ORIGINS_INTERVAL = 1

local balancers = {}
local site_origins   = {}
local upstream_servers = {}


-- local lock, err = lrucache.new(100)
-- if not lock then
--     error("failed to create the cache: " .. (err or "unknown"), 2)
-- end

function _M.add_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to add origin of the site '", host, "', origin already exists")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
    -- sync_lookup_origins()
end


function _M.del_origins(_, items)
    for _, host in ipairs(items) do
        if not origins[host] then
            log.warn("Failed to remove the origin server of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = nil
        balancers[host] = nil
        upstream_servers[host] = nil
        ::continue::
    end
    -- sync_lookup_origins()
end


function _M.update_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to update origin of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
    -- sync_lookup_origins()
end


function _M.full_sync_origins(_, items)
    local new_origins = {}
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins
        new_origins[host] = table_clone(origins)
    end
    site_origins = new_origins
end


-- local function parse_orgins(origins)
--     local servers = {
--         http  = tab_new(0, #origins),
--         https = tab_new(0, #origins)
--     }

--     for _, origin in ipairs(origins) do
--         local addr_http  = origin.host .. ":" .. origin.http_port
--         local addr_https = origin.host .. ":" .. origin.https_port
--         local weight = origin.weight > 0 and origin.weight or DEFAULT_SERVER_WEIGHT
--         servers.http[addr_http] = weight
--         servers.https[addr_https] = weight
--     end
--     return servers
-- end

local function balancer_reinit(host)
    local candidate = balancers[host]
    if not candidate then return end

    local servers = upstream_servers[host]
    if not servers then return end

    candidate:reinit(servers)
end



local function sync_lookup_origins()
    for host, _origins in pairs(site_origins) do

        local servers = {}
        for _, o in ipairs(_origins) do
            local ip_or_hostname, port, weight  = o.host, o.http_port, o.weight  -- TODO: http_port to port

            -- if not a valid ip, need to resolve host
            if not net.is_valid_ip(ip_or_hostname) then
                local addresses, _, err = dns.lookup(ip_or_hostname)
                if not addresses then
                    log.error("Could not resolve hostname '", ip_or_hostname, "'")
                    addresses = { ip_or_hostname }   -- TODO: make it better
                end

                for _, ip in ipairs(addresses) do
                    local id = ip .. ":" .. port

                    -- TODO: Duplicate case
                    servers[id] = weight
                end
            else
                -- ip case
                local id = ip_or_hostname .. ":" .. port
                servers[id] = weight
            end
        end

        upstream_servers[host] = servers

        -- reinit balancer make it effective
        -- log.debug("reinit balancer for host '", host, "'")
        balancer_reinit(host)
    end
end

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


-- because socket API disabled in the context of balancer phase, so use a timer to do it.
local function sync_lookup_origins_once()
    local ok, err = ngx.timer.at(0, sync_lookup_origins)
    if not ok then
        ngx.log(ngx.ERR, "failed to create timer: ", err)
    end
end


function _M.init_worker()
    sync_lookup_origins_once()

    log.debug("start timer for sync resolve origins")
    local ok, err = ngx.timer.every(SYNC_RESOLVE_ORIGINS_INTERVAL, sync_lookup_origins)
    if not ok then
        ngx.log(ngx.ERR, "failed to create timer: ", err)
    end
end


-- make sure that upstream servers is ready
function _M.access_phase(ctx)
    local host = ctx.host
    if not upstream_servers[host] then
        sync_lookup_origins()
        log.debug("upstream servers is not ready, synchronize immediately")
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
        log.error("failed to get balancer, ", err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end

    local peer = balancer:find()

    -- ok, err = balancer.set_current_peer(host, port)
    -- Domain names in host do not make sense. 
    local ok, err = ngx_balancer.set_current_peer(peer)
    if not ok then
        log.error("failed to setting current upstream peer '", peer, "', ", err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end
    log.debug("current upstream peer is '", peer, "'")
end


return _M