local require       = require
local log           = require("log")
local sample_log    = require("sample_log")
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

local CONFIG_SYNC_INTERVAL        = constants.CONFIG_SYNC_INTERVAL
local SAMPLE_LOG_SYNC_INTERVAL    = constants.SAMPLE_LOG_SYNC_INTERVAL

local _M = {
  _VERSION = 0.1
}

local function send_sample_logs()
    -- just need one worker
    if ngx_worker_id() == 0 then
        log.debug("start timer for send sample logs on worker ", tostring(ngx_worker_id()))
        local ok, err = every(SAMPLE_LOG_SYNC_INTERVAL, sample_log.iterator)
        assert(ok, "failed to setting up timer for save logs: " .. tostring(err))        
    end
end

local function sync_config()
    log.debug("start timer for sync configs on worker ", tostring(ngx_worker_id()))
    local ok, err = every(CONFIG_SYNC_INTERVAL, events.pop, ngx_worker_id())
    assert(ok, "failed to setting up timer for config sync: " .. tostring(err))
end


local function start()
    sync_config()
    send_sample_logs()
end


_M.init_worker = start


return _M