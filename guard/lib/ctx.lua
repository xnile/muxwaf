local require            = require
local balancer           = require("balancer")
local page_403           = require("page.403")
local page_410           = require("page.410")
local sites              = require("sites")
local ssl                = require("ssl")
local time               = require("time")
local log                = require("log")
local tablepool          = require("resty.tablepool")
local cjson              = require("cjson.safe")
local stringx            = require("utils.stringx")
local net                = require("utils.net")
local ngx                = ngx
local ngx_var            = ngx.var
local ngx_say            = ngx.say
local ngx_unescape_uri   = ngx.unescape_uri
local table_new          = table.new
local setmetatable       = setmetatable
local string_upper       = string.upper
local str_byte           = string.byte

local _M  = {
  _VERSION = 0.1,
}

local _mt = { __index = _M }

local HTTP_GONE              = ngx.HTTP_GONE
local HTTP_FORBIDDEN         = ngx.HTTP_FORBIDDEN
local DEFAULT_REAL_IP_HEADER = "X-Forwarded-For"



local ip_geo_searcher
local function get_ip_geo(ip)
    if not ip_geo_searcher then
        ip_geo_searcher = muxwaf.get_ip_geo_searcher()
    end
    local loc, err = ip_geo_searcher:search(ip)
    if err ~= nil then
      log.error("Failed to obtain location information for '", ip, "'")
    end
    return loc
end


local function get_real_client_ip(host, remote_addr)
  if not sites.is_real_ip_from_header(host) then
    return remote_addr
  end

  local real_ip_header = sites.get_real_ip_header(host)

  if not real_ip_header or real_ip_header == '' then
    real_ip_header = DEFAULT_REAL_IP_HEADER
  end


  local req_headers = ngx.req.get_headers()
  local raw_header_ip = req_headers[real_ip_header]
  if not raw_header_ip then
    log.warn("failed to get client ip from http header, '", string_upper(real_ip_header), "' header does not found, fallback to use remote_addr")
    return remote_addr
  end

  -- local xffs, err = ngx_re_split(raw_header_ip, ", ")
  -- if not xffs then
  --   return remote_addr
  -- end

  -- local real_client_ip = xffs[1]

  local real_client_ip = ''
  local idx = stringx.lfind_char(raw_header_ip, ',')
  if not idx then
    real_client_ip = raw_header_ip
  else
    for i = idx -1, 1, -1 do
      if str_byte(raw_header_ip, i) == str_byte(" ") then
        idx = idx - 1
      else
        break
      end
    end

    real_client_ip = raw_header_ip:sub(1, idx - 1)
  end

  if not net.is_valid_ip(real_client_ip) then
    log.warn("failed to get client ip from http header, '", real_client_ip, "' is invalid ip, fallback to use remote_addr")
    return remote_addr
  end

  return real_client_ip
end


local function say_block(ctx)
  -- metrics.incr_block_count()

  local request_id = ctx.request_id
  local page_403 = page_403
  page_403[2] = request_id
  ngx.header["Content-Type"] = 'text/html'  -- should before ngx.status
  ngx.status = HTTP_FORBIDDEN
  ngx_say(page_403)
  ngx.exit(ngx.status)
end


local function say_410()
    ngx.header["Content-Type"] = "text/html"
    ngx.status = HTTP_GONE
    ngx_say(page_410)
    ngx.exit(ngx.status)
end


local function decode_url(url)
  return ngx_unescape_uri(url)
end


local function set_vars(ctx)
  ngx_var.x_real_ip = ctx.real_client_ip

  -- should before balance phase, otherwise it will not take effect
  if ctx.upstream_scheme and ctx.upstream_scheme ~= "" then
    ngx_var.upstream_scheme = ctx.upstream_scheme
    log.debug("set the host for requesting the origin site use '", ctx.upstream_scheme, "'")
  end

  -- set origin host
  if ctx.upstream_host and ctx.upstream_host ~= "" then
    ngx_var.upstream_host = ctx.upstream_host
    log.debug("set the host for requesting the origin site to '", ctx.upstream_host, "'")
  end
end


-- function _M.set_param(self, param)
--   self.param = param
--   return
-- end

function _M.new()
  local now = time.now()
  local ctx = tablepool.fetch("pool_ctx", 0, 20)
  -- local ctx = table_new(0, 20)
  -- ctx.var = vars.new()

  ctx.host               = ngx_var.host
  ctx.scheme             = ngx_var.scheme
  ctx.request_id         = ngx_var.request_id
  ctx.remote_addr        = ngx_var.remote_addr
  ctx.request_method     = ngx_var.request_method
  ctx.request_uri        = ngx_var.request_uri
  ctx.request_path       = ngx_var.uri
  ctx.server_port        = ngx_var.server_port

  ctx.sample_log = {}
  ctx.waf_start_time = now
  ctx.site_id = sites.get_site_id(ctx.host)
  ctx.real_client_ip = get_real_client_ip(ctx.host, ctx.remote_addr)  
  ctx.upstream_host   = balancer.get_origin_host(ctx.host)
  ctx.ip_geo = get_ip_geo(ctx.real_client_ip)
  -- ctx.unescape_uri = decode_url(ctx.request_uri)

  ctx.upstream_scheme, ctx.upstream_server = balancer.get_origin_peer_and_protocol(ctx.host, ctx.scheme)

  ctx.say_410 = say_410
  ctx.say_block = say_block
  ctx.encode = cjson.encode
  ctx.decode = cjson.decode

  set_vars(ctx)

  return setmetatable(ctx, _mt)
end

return _M