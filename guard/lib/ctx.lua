local require            = require
local lrucache           = require "resty.lrucache.pureffi"
local page_403           = require("page.403")
local page_410           = require("page.410")
local sites              = require("sites")
local ssl                = require("ssl")
local net                = require("utils.net")
local vars               = require("vars")
local time               = require("time")
local log                = require("log")
-- local metrics            = require("metrics")
local ngx_re_split       = require("ngx.re").split
local tablepool          = require("resty.tablepool")
local cjson              = require("cjson.safe")
local stringx            = require("utils.stringx")
local ngx                = ngx
local ngx_say            = ngx.say
local ngx_worker_id      = ngx.worker.id
-- local str_sub            = ngx.re.sub
local ngx_unescape_uri   = ngx.unescape_uri
local req_read_body      = ngx.req.read_body
local req_get_body_data  = ngx.req.get_body_data
local req_get_body_file  = ngx.req.get_body_file
local error              = error
local io_open            = io.open
local io_close           = io.close
local table_new          = table.new
local setmetatable       = setmetatable
local string_upper       = string.upper
local JSON_NULL          = cjson.null


local IP_LOC_CACHE_SIZE = 1000

local HTTP_GONE              = ngx.HTTP_GONE
local HTTP_FORBIDDEN         = ngx.HTTP_FORBIDDEN
local HTTP_NOT_FOUND         = ngx.HTTP_NOT_FOUND
local DEFAULT_REAL_IP_HEADER = "X-Forwarded-For"

local _M  = {
  _VERSION = 0.1,
}

local _mt = { __index = _M }

local ip_loc_cache, err = lrucache.new(IP_LOC_CACHE_SIZE)
if not ip_loc_cache then
    error("failed to create the cache: " .. (err or "unknown"), 2)
end

local ipdb
local function get_ip_location(ip)
    if not ipdb then
        ipdb = muxwaf.get_ipdb()
    end
    local location = ip_loc_cache:get(ip)
    if not location then
        -- {"country_name":"中国","region_name":"北京","city_name":"北京"}
        location = ipdb.ipip:find(ip, "CN")
        ip_loc_cache:set(ip, location, 3600)
    end
    return location
end

local function get_body_data()
    req_read_body()
    local body = req_get_body_data()

    if not body then
      -- request body might've been written to tmp file if body > client_body_buffer_size
      local file_name = req_get_body_file()
      if not file_name then
        return nil
      end

      local fd = io_open(file_name, "rb")
      if not fd then
        return nil
      end

      body = fd:read("*all")
      io_close(fd)
    end

    return body
end


local function get_and_decode_body_data()
  local raw = get_body_data()
  if not raw then
    return nil, "the data passed cannot be empty"
  end

  local data, err = cjson.decode(raw)
  if not data then
    return nil, err
  end
  return data, nil
end

local function get_real_client_ip(host, remote_addr)
  if not sites.is_real_ip_from_header(host) then
    return remote_addr
  end

  local real_ip_header = sites.get_real_ip_header(host)

  if real_ip_header == '' then
    real_ip_header = DEFAULT_REAL_IP_HEADER
  end


  local raw_header_ip = ngx.req.get_headers()[real_ip_header]
  if not raw_header_ip then
    log.warn("failed to get client ip from http header, \"", string_upper(real_ip_header), "\" header does not found, fallback to use remote_addr")
    return remote_addr
  end

  -- local xffs, err = ngx_re_split(raw_header_ip, ", ")
  -- if not xffs then
  --   return remote_addr
  -- end

  -- local real_client_ip = xffs[1]

  local real_client_ip = ''
  local idx = stringx.rfind_char(raw_header_ip, ',')
  if not idx then
    real_client_ip = raw_header_ip
  else
    real_client_ip = raw_header_ip:sub(1, idx-1)
  end

  if not net.is_valid_ip(real_client_ip) then
    log.warn("failed to get client ip from http header: ip \"", real_client_ip, "\" is invalid, fallback to use remote_addr")
    return remote_addr
  end

  return real_client_ip
end

local function say_block(ctx)
  -- metrics.incr_block_count()

  local request_id = ctx.var.request_id
  local page_403 = page_403
  page_403[2] = request_id
  ngx.header["Content-Type"] = 'text/html'  -- should before ngx.status
  ngx.status = HTTP_FORBIDDEN
  ngx_say(page_403)
  ngx.exit(ngx.status)
end

local function say_ok()
    ngx.header["Content-Type"] = 'application/json'
    return ngx_say(cjson.encode({
      code = 0,
      msg  = "Success",
      data = JSON_NULL,
    })) 
end


local function say_err(code, msg)
  ngx.header["Content-Type"] = 'application/json'
  return ngx_say(cjson.encode({
    code = code,
    msg  = msg,
    data = JSON_NULL
  }))
end

local function say_404()
  ngx.header["Content-Type"] = 'application/json'
  ngx.status= HTTP_NOT_FOUND
  return ngx_say("404 page not found")
end

local function say_410()
    ngx.header["Content-Type"] = "text/html"
    ngx.status = HTTP_GONE
    ngx_say(page_410)
    ngx.exit(ngx.status)
end

local function say_json(data)
  ngx.header["Content-Type"] = 'application/json'
  return ngx_say(cjson.encode(data))
end

local function decode_url(url)
  return ngx_unescape_uri(url)
end


function _M.set_param(self, param)
  self.param = param
  return
end

function _M.new()
  local ctx = tablepool.fetch("pool_ctx", 0, 30)
  -- local ctx = table_new(0, 25)
  ctx.var = vars.new()
  ctx.param = {} -- Parameters in path
  ctx.say_ok = say_ok
  ctx.say_err = say_err
  ctx.say_404 = say_404
  ctx.say_410 = say_410
  ctx.say_json = say_json
  ctx.say_block = say_block
  ctx.encode = cjson.encode
  ctx.decode = cjson.decode
  ctx.get_body_data = get_body_data
  ctx.get_and_decode_body_data = get_and_decode_body_data  
  ctx.waf_start_time = time.now()
  ctx.worker_id = ngx_worker_id()
  ctx.request_path = ctx.var.uri
  ctx.request_url = decode_url(ctx.var.request_uri)
  ctx.site_id = sites.get_site_id(ctx.var.host)
  ctx.real_client_ip = get_real_client_ip(ctx.var.host, ctx.var.remote_addr)  
  ctx.upstream_scheme = sites.get_origin_protocol(ctx.var.host, ctx.var.scheme)
  ctx.location = get_ip_location(ctx.real_client_ip)
  
  --TODO: move to set func
  ctx.var.x_real_ip = ctx.real_client_ip
  ctx.var.upstream_scheme = ctx.upstream_scheme
  ctx.sample_log = {}
  ctx.blocked = false
  return setmetatable(ctx, _mt)
end

return _M