local require   = require
local cjson     = require("cjson.safe")
local tree      = require("tree")
local table_new = table.new
local ngx       = ngx
local ngx_var   = ngx.var
local ngx_say   = ngx.say
local ipairs    = ipairs
local error     = error
local io_open            = io.open
local io_close           = io.close
local req_read_body      = ngx.req.read_body
local req_get_body_data  = ngx.req.get_body_data
local req_get_body_file  = ngx.req.get_body_file
local JSON_NULL          = cjson.null

local _M = {
    _VERSION = 0.1
}

local HTTP_NOT_FOUND = ngx.HTTP_NOT_FOUND
local METHODS = {"GET", "POST", "DELETE", "PATCH", "PUT", "OPTIONS", "HEAD"}


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

local function say_json(data)
  ngx.header["Content-Type"] = 'application/json'
  return ngx_say(cjson.encode(data))
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


local function start(self)
    local uri = ngx_var.uri
    local method = ngx_var.request_method

    if not self.mux[method] then
        return say_404()
    end

    local handler, param = self.mux[method]:get(uri)
    if not handler then
        return say_404()
    end

    local ctx = {
        param = param,
        encode = cjson.encode,
        decode = cjson.decode,
        say_err = say_err,
        say_ok = say_ok,
        say_json = say_json,
        get_and_decode_body_data = get_and_decode_body_data,
    }

    handler(ctx)
end


function _M.new(prefix)
    local mux = table_new(0, #METHODS)
    for _, method in ipairs(METHODS) do
        mux[method] = tree.new()
    end

    local self = {
        mux = mux,
        start = start,
    }

    return setmetatable(self, {
        __index = function(self, key)
            return function(self, path, handler)
                if not self or not path or not handler then
                    error("missing parameters", 2)
                    return
                end

                if not path or path == '' then
                    error("path cannot be empty", 2)
                    return
                end

                if not handler then
                    error("handler cannot be empty", 2)
                    return
                end

                if not self.mux[key] then
                    error(key .. " method is not supported", 2)
                    return
                end

                self.mux[key]:insert(prefix .. path, handler)
            end
        end
    })
end

return _M