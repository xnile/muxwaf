-- TODO: add cache
local require  = require
local log      = require("log")
local resolver = require("resty.dns.resolver")
local ngx_re_match  = require("ngx.re").match
local ngx_re_split  = require("ngx.re").split
local io_open       = io.open
local io_close      = io.close
local string_format = string.format

local _M = {
    _VERSION = 0.1
}

-- maximum value according to https://tools.ietf.org/html/rfc2181
-- take frome https://github.com/kubernetes/ingress-nginx/blob/main/rootfs/etc/nginx/lua/util/dns.lua
local MAX_TTL = 2147483647
local DEFAULT_RESOLV_CONF = "/etc/resolv.conf"

local parse_resolve_conf = function(filename)
    local nameservers = {}

    local f, err = io_open(filename or DEFAULT_RESOLV_CONF)
    if not f then
        error(err, 2)
    end

    for line in f:lines() do
        local m, _ = ngx.re.match(line, "(^\\s?(#|;))")
        if m then
            goto continue
        end

        local seg, _ = ngx_re_split(line, "\\s+")

        local option, value = seg[1], seg[2]

        if option == "nameserver" then
            nameservers[#nameservers +1] = value
        else
            -- TODO "search,ndots"
        end

        ::continue::
    end
    f:close()
    return nameservers    
end

local function get_ns_servers()
    return parse_resolve_conf(DEFAULT_RESOLV_CONF)
end

function _M.lookup(qname)
    local addresses = {}

    local r, err = resolver:new{
        nameservers = get_ns_servers(),
        retrans = 2,
        timeout = 1000,  -- 1 sec
    }

    if not r then
        log.error("failed to instantiate the resolver: ", err)
        return { qname }
    end

    local answers, err = r:query(qname, { qtype = resolver.TYPE_A })
        if not answers then
        log.error("failed to query the DNS server: ", err)
        return { qname }
    end
 
    if answers.errcode then
        log.error(string_format("server returned error code: %s: %s", answers.errcode, answers.errstr))
        return { qname }
    end

    for i, ans in ipairs(answers) do
        addresses[#addresses +1] = ans.address
    end
    
    return addresses

end

return _M