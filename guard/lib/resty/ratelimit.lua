-- Fixed Window algorithm

local ngx = ngx
local ngx_shared =  ngx.shared
local setmetatable = setmetatable
local error = error

local _M = {}
local _mt = { __index = _M }

function _M.new(dict_name)
  local dict = ngx_shared[dict_name]
  if not dict then
    return error("lua_shared_dict '" .. dict_name .. "' not found", 2)
  end

  local self = {
    dict = dict
  }

  return setmetatable(self, _mt)
end

function _M.incomming(self, key, limit, window)
  local dict = self.dict
  local remaining, err = dict:incr(key, -1, limit, window)
  if not remaining then
    return nil, err
  end

  if remaining < 0 then
    return nil, "rejected"
  end

  return 0, remaining
end


function _M.flush_all(self)
  local shm = self.dict
  shm:flush_all()
  shm:flush_expired()
end

return _M
