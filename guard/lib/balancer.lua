local require       = require
local sites         = require("sites")
local log           = require("log")
local ngx_balancer  = require("ngx.balancer")
local roundrobin    = require "resty.balancer.roundrobin"
local cjson         = require("cjson.safe")
local ngx           = ngx
local ngx_exit      = ngx.exit
local ipairs        = ipairs
local tab_new       = table.new
local string_format = string.format

local HTTP_BAD_GATEWAY          = ngx.HTTP_BAD_GATEWAY
local DEFAULT_UPSTREAM_SERVER   = "127.0.0.1:9000"

local upstreams = {}

local _M = { 
  _VERSION = 0.1
}

local function parse_orgins(origins)
    local len = #origins
    local default_weight = 10
    local servers = tab_new(0, 2)
    servers.http  = tab_new(0, len)
    servers.https = tab_new(0, len)

    for _, origin in ipairs(origins) do
        local addr_http  = origin.host .. ":" .. origin.http_port
        local addr_https = origin.host .. ":" .. origin.https_port
        local weight = origin.weight > 0 and origin.weight or default_weight
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

    local servers = parse_orgins(origins)
    local rr = roundrobin:new(servers[upstream_protocol])
    local peer = rr:find()

    if not peer then
        return nil
    end

    return peer
end


local function get_upstream(host, protocol)
    return upstreams[host][protocol]
end


function _M.balance(self, ctx)
    ngx_balancer.set_more_tries(1)

    local peer = pick_upstream_peer(ctx.var.host, ctx.upstream_scheme)
    if not peer then
        return ngx_exit(HTTP_BAD_GATEWAY)
    end

    log.info("[peer->", peer, "]")

    local ok, err = ngx_balancer.set_current_peer(peer)
    if not ok then
        log.error(string_format("failed to setting current upstream peer %s : %s", peer, err))
    end
end


return _M