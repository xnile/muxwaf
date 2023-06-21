local require       = require
local log           = require("log")
local lock          = require("lock")
local ngx_balancer  = require("ngx.balancer")
local roundrobin    = require("resty.balancer.roundrobin")
-- local lrucache      = require("resty.lrucache")
local cjson         = require("cjson.safe")
local net           = require("utils.net")
local dns           = require("utils.net.dns")
local stringx       = require("utils.stringx")
local table_clone   = require("table.clone")
local ngx           = ngx
local ngx_exit      = ngx.exit
local ngx_get_phase = ngx.get_phase
local tab_new       = table.new
local table_concat  = table.concat
local ipairs        = ipairs

local _M = { 
  _VERSION = 0.1
}

local HTTP_BAD_GATEWAY          = ngx.HTTP_BAD_GATEWAY
local HTTP_GATEWAY_TIMEOUT      = ngx.HTTP_GATEWAY_TIMEOUT
local SYNC_LOOKUP_ORIGINS_INTERVAL = 5
local MAXIMUM_TTL_VALUE = 2147483647

local balancers = {}
local site_origins   = {}
local upstream_servers = {}

local dns_lookup_lock, err = lock.new(100)
if not dns_lookup_lock then
    error("failed to create the cache: " .. (err or "unknown"), 2)
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
    local servers = upstream_servers[host]
    if not servers then
        log.error("upstream servers for the site '", host, "' were not found")
        return
    end

    if not balancers[host] then
        balancers[host] = roundrobin:new(servers)
        return
    end

    balancers[host]:reinit(servers)
end


local function delete_upstream_servers_and_balancer(host)
    balancers[host] = nil
    upstream_servers[host] = nil
    dns_lookup_lock:unlock(host)
end


local function resolve_origins_and_reinit_balancer(host, force)
    local origins = site_origins[host] or {}
    local servers = {}
    local ttl = MAXIMUM_TTL_VALUE

    if force then
        dns_lookup_lock:unlock(host)
    end

    local ok = dns_lookup_lock:lock(host)
    if not ok then
        log.debug("skip updating the origin for '", host, "'")
        return
    end

    for _, o in ipairs(origins) do
        -- TODO: rename host to addr, http_port to port
        local addr, port, weight, protocol, origin_type = o.host, o.http_port, (o.weight or 100), (o.protocol or "http"), (o.origin_type or "")
        if origin_type == "" then
            origin_type = net.is_valid_ip(addr) and "ip" or "domain"
        end

        if origin_type == "ip" then
            local id = table_concat({protocol, addr, port}, ":")
            servers[id] = weight
        elseif origin_type == "domain" then
            local addresses, min_ttl, err = dns.lookup(addr)
            if not addresses then
                log.error("could not resolve host '", addr, "'")
                addresses = {addr}
            end

            for _, ip in ipairs(addresses) do
                local id = table_concat({protocol, ip, port}, ":")
                servers[id] = weight
            end

            if min_ttl < ttl then
                ttl = min_ttl
            end
        end
    end

    upstream_servers[host] = servers

    balancer_reinit(host)

    -- has domain origin
    if ttl < MAXIMUM_TTL_VALUE then
        log.debug("'", host, "' origin is locked for updates for ", ttl, " seconds.")
        dns_lookup_lock:unlock(host)
        dns_lookup_lock:lock(host, ttl)
    end

end


-- local function sync_lookup_origins()
--     for host, _origins in pairs(site_origins) do

--         local servers = {}
--         for _, o in ipairs(_origins) do
--             local ip_or_domain, port, weight, protocol = o.host, o.http_port, o.weight, o.protocol  -- TODO: http_port to port
--             protocol = protocol or "http"

--             -- if not a valid ip, need to resolve host
--             if not net.is_valid_ip(ip_or_domain) then
--                 local addresses, _, err = dns.lookup(ip_or_domain)
--                 if not addresses then
--                     log.error("Could not resolve hostname '", ip_or_domain, "'")
--                     addresses = { ip_or_domain }   -- TODO: make it better
--                 end

--                 for _, ip in ipairs(addresses) do
--                     local id = table_concat({protocol, ip, port}, ":")

--                     -- TODO: Duplicate case
--                     servers[id] = weight
--                 end
--             else
--                 -- ip case
--                 local id = table_concat({protocol, ip_or_domain, port}, ":")
--                 servers[id] = weight
--             end
--         end

--         upstream_servers[host] = servers

--         -- reinit balancer make it effective
--         -- log.debug("reinit balancer for host '", host, "'")
--         balancer_reinit(host)
--     end
-- end


local function sync_resolve_origins_and_reinit_balancer()
    for host, _ in pairs(site_origins) do
        resolve_origins_and_reinit_balancer(host)
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


-- return protocol, peer
function _M.get_origin_peer_and_protocol(host, scheme)
    local balancer = get_balancer(host)
    if not balancer then return nil, nil end
    local raw_peer = balancer:find()

    local idx = stringx.lfind_char(raw_peer, ":")
    if not idx then
        return nil, nil
    end

    local peer, protocol = raw_peer:sub(1, idx - 1), raw_peer:sub(idx + 1)
    if protocol == "follow" then   -- TODO:
        log.error("Protocol following is currently not supported.")
        protocol = "http"
    elseif protocol ~= "http"  or protocol ~= "https" then
        log.error("Failed to obtain the origin protocol.")
        protocol = "http"
    end
    
    return peer, protocol
end



function _M.add_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to add origin of the site '", host, "', origin already exists")
            goto continue
        end

        site_origins[host] = table_clone(origins)

        -- init_worker phase, sites:full_sync -> sites:add() -> this
        -- TODO: better it
        if ngx_get_phase() ~= "init_worker" then
            resolve_origins_and_reinit_balancer(host)
        end

        ::continue::
    end
end


function _M.del_origins(_, items)
    for _, host in ipairs(items) do
        if not origins[host] then
            log.warn("Failed to remove the origin server of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = nil

        -- init_worker phase, sites:full_sync -> sites:del() -> this
        -- TODO: better it
        if ngx_get_phase() ~= "init_worker" then
            resolve_origins_and_reinit_balancer(host)
        end

        ::continue::
    end
end


function _M.update_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to update origin of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = table_clone(origins)

        resolve_origins_and_reinit_balancer(host, true)
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


function _M.init_worker()
    -- synchronize immediately.
    -- because socket API disabled in the context of balancer phase, so use a timer to do it.
    local ok, err = ngx.timer.at(0, sync_resolve_origins_and_reinit_balancer)
    if not ok then
        ngx.log(ngx.ERR, "failed to create timer: ", err)
    end

    log.debug("start timer for sync resolve origins")
    local ok, err = ngx.timer.every(SYNC_LOOKUP_ORIGINS_INTERVAL, sync_resolve_origins_and_reinit_balancer)
    if not ok then
        ngx.log(ngx.ERR, "failed to create timer: ", err)
    end
end


-- make sure that upstream servers is ready
function _M.access_phase(ctx)
    -- local host = ctx.host
    -- if not upstream_servers[host] then
    --     sync_resolve_origins_and_reinit_balancer()
    --     log.debug("upstream servers is not ready, synchronize immediately")
    -- end
end


function _M.balance(ctx)
    -- local host = ctx.host

    -- if not host then
    --     log.error("failed to get host")
    --     return ngx_exit(HTTP_GATEWAY_TIMEOUT)
    -- end

    -- ngx_balancer.set_more_tries(1)

    -- local balancer, err = get_balancer(host)
    -- if not balancer then
    --     log.error("failed to get balancer, ", err)
    --     return ngx_exit(HTTP_BAD_GATEWAY)
    -- end

    -- local peer = balancer:find()
    local peer = ctx.upstream_server
    if peer == nil then
        log.error("upstream peer is nil")
        return ngx_exit(HTTP_BAD_GATEWAY)
    end

    log.debug("current upstream peer is '", peer, "'")

    -- ok, err = balancer.set_current_peer(host, port)
    -- Domain names in host do not make sense.
    local ok, err = ngx_balancer.set_current_peer(peer)
    if not ok then
        log.error("failed to setting current upstream peer '", peer, "', ", err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end
end


return _M