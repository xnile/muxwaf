local require       = require
local log           = require("log")
local events        = require("events")
local metrics       = require("metrics")
local constants     = require("constants")
local cjson         = require("cjson.safe")
local ngx           = ngx
local every         = ngx.timer.every
local ngx_worker_id = ngx.worker.id
local io_open       = io.open
local tostring      = tostring
local assert        = assert

local CONFIG_SYNC_INTERVAL  = constants.CONFIG_SYNC_INTERVAL
local LOG_SYNC_INTERVAL     = constants.LOG_SYNC_INTERVAL
local CALC_QPS_INTERVAL     = constants.CALC_QPS_INTERVAL

local _M = {
  _VERSION = 0.1
}

local function save_logs()
    -- just need one worker
    if ngx_worker_id() == 0 then
        local ok, err = every(LOG_SYNC_INTERVAL, log.iterator)
        assert(ok, "Failed to setting up timer for save logs: " .. tostring(err))        
    end
end

local function calc_qps()
    -- just need one worker
    if ngx_worker_id() == 0 then
        local ok, err = every(CALC_QPS_INTERVAL, metrics.calc_qps)
        assert(ok, "Failed to setting up timer for calculate qps: " .. tostring(err))        
    end
end


local function sync_config()
    local ok, err = every(CONFIG_SYNC_INTERVAL, events.pop, ngx_worker_id())
    assert(ok, "Failed to setting up timer for config sync: " .. tostring(err))
end


function _M.run(_)
    sync_config()
    save_logs()
    calc_qps()
end

return _M