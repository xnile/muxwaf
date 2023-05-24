local require       = require
local sites         = require("sites")
local log           = require("log")
local ngx_balancer  = require("ngx.balancer")
local roundrobin    = require("resty.balancer.roundrobin")
local lrucache      = require("resty.lrucache")
local cjson         = require("cjson.safe")
local ngx           = ngx
local ngx_exit      = ngx.exit
local ipairs        = ipairs
local tab_new       = table.new

local HTTP_BAD_GATEWAY          = ngx.HTTP_BAD_GATEWAY
local HTTP_GATEWAY_TIMEOUT      = ngx.HTTP_GATEWAY_TIMEOUT

local _M = { 
  _VERSION = 0.1
}

local upstreams = {}

local CACHE_TTL  = 60
local CACHE_SIZE = 100
local DEFAULT_SERVER_WEIGHT = 100

local cache, err = lrucache.new(CACHE_SIZE * 2)
if not cache then
    error("failed to create the cache: " .. (err or "unknown"), 2)
end


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


local function pick_upstream_peer(host, upstream_protocol)
    local origins = sites.get_site_origins(host)
    if not origins then
        return nil
    end

    local servers = cache:get(host)
    if not servers then
        servers = parse_orgins(origins)
        cache:set(host, servers, CACHE_TTL)
    end

    local candidate = servers[upstream_protocol]
    if not candidate then
        return nil
    end

    local rr = roundrobin:new(candidate)
    local peer = rr:find()

    if not peer then
        return nil
    end

    return peer
end


local function get_upstream(host, protocol)
    return upstreams[host][protocol]
end


function _M.balance(ctx)
    if not ctx.host then
        log.error("failed to get host")
        return ngx_exit(HTTP_GATEWAY_TIMEOUT)
    end

    ngx_balancer.set_more_tries(1)

    local peer = pick_upstream_peer(ctx.host, ctx.upstream_scheme)
    if not peer then
        return ngx_exit(HTTP_BAD_GATEWAY)
    end

    local ok, err = ngx_balancer.set_current_peer(peer)
    if not ok then
        log.error("failed to setting current upstream peer \"", peer, "\", ", err)
        return ngx_exit(HTTP_BAD_GATEWAY)
    end
    log.debug("current upstream peer \"", peer, "\"")
end


return _M