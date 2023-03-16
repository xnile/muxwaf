local require            = require
local page_403           = require("page.403")
local page_410           = require("page.410")
local sites              = require("sites")
local ssl                = require("ssl")
local net                = require("utils.net")
local vars               = require("vars")
local time               = require("time")
local log                = require("log")
local metrics            = require("metrics")
local ngx_re_split       = require("ngx.re").split
local tablepool          = require("resty.tablepool")
local cjson              = require("cjson.safe")
local ngx                = ngx
local ngx_say            = ngx.say
local ngx_worker_id      = ngx.worker.id
local str_sub            = ngx.re.sub
local ngx_unescape_uri   = ngx.unescape_uri
local req_read_body      = ngx.req.read_body
local req_get_body_data  = ngx.req.get_body_data
local req_get_body_file  = ngx.req.get_body_file
local error              = error
local io_open            = io.open
local io_close           = io.close
local table_new          = table.new
local setmetatable       = setmetatable
local string_format      = string.format
local string_upper       = string.upper
local JSON_NULL          = cjson.null

local HTTP_GONE              = ngx.HTTP_GONE
local HTTP_FORBIDDEN         = ngx.HTTP_FORBIDDEN
local HTTP_NOT_FOUND         = ngx.HTTP_NOT_FOUND
local DEFAULT_REAL_IP_HEADER = "X-Forwarded-For"

local _M  = {
  _VERSION = 0.1,
}


local _mt = { __index = _M }

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
    log.error(string_format("failed to get ip from http header: '%s' header does not found, fallback to use remote_addr", string_upper(real_ip_header)))
    return remote_addr
  end

  local xffs, err = ngx_re_split(raw_header_ip, ", ")
  if not xffs then
    return remote_addr
  end

  local client_real_ip = xffs[1]

  if not net.is_valid_ip(client_real_ip) then
    log.error(string_format('failed to get ip from http header: ip '%s' is invalid', client_real_ip))
    return remote_addr
  end

  return client_real_ip
end

local function say_block(ctx)
  metrics.incr_block_count()

  local request_id = ctx.var.request_id
  local rendered_page, _, err = str_sub(page_403, "{{REQUEST_ID}}", request_id)
  if not rendered_page then
    log.error(string_format("failed to render page: %s", err))
    return ngx_exit(HTTP_FORBIDDEN)
  end
  ngx.header["Content-Type"] = 'text/html'  -- should before ngx.status
  ngx.status = HTTP_FORBIDDEN
  ngx_say(rendered_page)
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
  local ctx = tablepool.fetch("pool_ctx", 0, 25)
  -- local ctx = table_new(0, 25)
  ctx.var = vars
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
  ctx.start_time = time.now()
  ctx.worker_id = ngx_worker_id()
  ctx.request_path = ctx.var.uri
  ctx.request_url = decode_url(ctx.var.request_uri)
  ctx.site_id = sites.get_site_id(ctx.var.host)
  ctx.real_client_ip = get_real_client_ip(ctx.var.host, ctx.var.remote_addr)  
  ctx.upstream_scheme = sites.get_origin_protocol(ctx.var.host, ctx.var.scheme)
  
  --TODO: move to set func
  ctx.var.upstream_x_real_ip = ctx.real_client_ip
  ctx.var.upstream_scheme = ctx.upstream_scheme
  return setmetatable(ctx, _mt)
end

return _M