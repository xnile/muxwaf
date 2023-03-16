local require = require
local configs = require("configs")
local cjson   = require("cjson.safe")
local ngx     = ngx
local ngx_worker_count = ngx.worker.count
local assert  = assert

local _dict_name = require("constants").DICTS.EVENTS
local shm = ngx.shared[_dict_name]

local _M = {
    _VERSION = 0.1
}


local function handle(event)
  local r, err = cjson.decode(event)
  if not r then
    log.error("could not parse event: ", err)
    return
  end

  local configType, operation, data = r.configType, r.operation, r.data
  configs:sync(configType, operation, data)
end


function _M.send(_, event)
  for i = 0, ngx_worker_count() -1 do
    local key = "pid:" .. i
    shm:rpush(key, event)
  end
end


function _M.pop(self, worker_pid)
  local key = "pid:" .. worker_pid

  local len, err = shm:llen(key)
  if len == 0 then
    return
  end

  for i = 1, len do
    local event = assert(shm:lpop(key))
    handle(event)
  end
end


return _M