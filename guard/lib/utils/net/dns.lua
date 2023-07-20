-- Modified from https://github.com/kubernetes/ingress-nginx/blob/main/rootfs/etc/nginx/lua/util/dns.lua

local resolver = require("resty.dns.resolver")
local lrucache = require("resty.lrucache")
-- local resolv_conf = require("util.resolv_conf")
local resolv_conf = require("utils.net.resolv_conf")

local ngx_log = ngx.log
local ngx_INFO = ngx.INFO
-- local ngx_ERR = ngx.ERR
local ngx_WARN = ngx.WARN
local string_format = string.format
local table_concat = table.concat
local table_insert = table.insert
local ipairs = ipairs
local tostring = tostring

local _M = {}
local CACHE_SIZE = 1000
-- maximum value according to https://tools.ietf.org/html/rfc2181
local MAXIMUM_TTL_VALUE = 2147483647
-- for every host we will try two queries for the following types with the order set here
local QTYPES_TO_CHECK = { resolver.TYPE_A, resolver.TYPE_AAAA }


local cache
do
  local err
  cache, err = lrucache.new(CACHE_SIZE)
  if not cache then
    return error("failed to create the cache: " .. (err or "unknown"))
  end
end

local function cache_set(host, addresses, ttl)
  cache:set(host, addresses, ttl)
  ngx_log(ngx_INFO, string_format("cache set for '%s' with value of [%s] and ttl of %s.",
    host, table_concat(addresses, ", "), ttl))
end

local function is_fully_qualified(host)
  return host:sub(-1) == "."
end

local function a_records_and_min_ttl(answers)
  local addresses = {}
  local ttl = MAXIMUM_TTL_VALUE -- maximum value according to https://tools.ietf.org/html/rfc2181

  for _, ans in ipairs(answers) do
    if ans.address then
      table_insert(addresses, ans.address)
      if ans.ttl < ttl then
        ttl = ans.ttl
      end
    end
  end

  return addresses, ttl
end

local function resolve_host_for_qtype(r, host, qtype)
  local answers, err = r:query(host, { qtype = qtype }, {})
  if not answers then
    return nil, -1, err
  end

  if answers.errcode then
    return nil, -1, string_format("server returned error code: %s: %s",
      answers.errcode, answers.errstr)
  end

  local addresses, min_ttl = a_records_and_min_ttl(answers)
  if #addresses == 0 then
    local msg = "no A record resolved"
    if qtype == resolver.TYPE_AAAA then msg = "no AAAA record resolved" end
    return nil, -1, msg
  end

  return addresses, min_ttl, nil
end

local function resolve_host(r, host)
  local dns_errors = {}

  for _, qtype in ipairs(QTYPES_TO_CHECK) do
    local addresses, ttl, err = resolve_host_for_qtype(r, host, qtype)
    if addresses and #addresses > 0 then
      return addresses, ttl, nil
    end
    table_insert(dns_errors, tostring(err))
  end

  return nil, nil, dns_errors
end

function _M.lookup(host)
  local cached_addresses = cache:get(host)
  if cached_addresses then
    return cached_addresses, -1, nil
  end

  local r, err = resolver:new{
    nameservers = resolv_conf.nameservers,
    retrans = 5,  -- 5 retransmissions on receive timeout
    timeout = 2000,  -- 2 sec
    no_random = true, -- always start with first nameserver
  }

  if not r then
    return nil, -1, string_format("failed to instantiate the resolver: %s", err)
  end

  local addresses, min_ttl, dns_errors
  addresses, min_ttl, dns_errors = resolve_host(r, host)
  if addresses then
    cache_set(host, addresses, min_ttl)
    return addresses, min_ttl, nil
  end


  local _, stale_addresses = cache:get(host)
  if stale_addresses then
      ngx_log(ngx_WARN, "failed to query the DNS server for ", host, ":\n", table_concat(dns_errors, "\n"),
       ", fallback to using the previous DNS resolution result")
      return  stale_addresses, 0, nil
  end

  return nil, -1, "failed to query the DNS server for " .. host .. ":\n" .. table_concat(dns_errors, "\n")
end

setmetatable(_M, {__index = { _cache = cache }})

return _M