local require     = require
local constants   = require("constants")
local utils       = require("utils")
local log         = require("log")
local ffi         = require "ffi"
local C           = ffi.C
local tonumber    = tonumber
local tostring    = tostring
local math        = math
local table_new   = table.new
local ngx_shared  = ngx.shared

local _M = {
    _VERSION = 0.1
}

local ditcs       = constants.DICTS
local shm_metrics = ngx_shared[ditcs.METRICS]
local CALC_QPS_INTERVAL = constants.CALC_QPS_INTERVAL


ffi.cdef[[
    uint64_t *ngx_stat_active;
    uint64_t *ngx_stat_reading;
    uint64_t *ngx_stat_writing;
    uint64_t *ngx_stat_waiting;
    uint64_t *ngx_stat_requests;
    uint64_t *ngx_stat_accepted;
    uint64_t *ngx_stat_handled;
]]


local function get_nginx_status()
    return {
        connections_active   = tonumber(C.ngx_stat_active[0]),
        connections_reading  = tonumber(C.ngx_stat_reading[0]),
        connections_writing  = tonumber(C.ngx_stat_writing[0]),
        connections_waiting  = tonumber(C.ngx_stat_waiting[0]),
        connections_accepted = tonumber(C.ngx_stat_accepted[0]),
        connections_handled  = tonumber(C.ngx_stat_handled[0]),
        total_requests       = tonumber(C.ngx_stat_requests[0])
    }
end

local function get_shm_status()
    local status = table_new(0, 3)
    for _, dict in pairs(ditcs) do
        local shm = ngx_shared[dict]
        status[dict] = {
            capacity = utils.format_capacity(shm:capacity()),
            free     = utils.format_capacity(shm:free_space()),
        }
    end
    return status
end


local function get_block_count()
    local count, err = shm_metrics:get("block_count")
    if err then
        log.error("Failed to get the number of blocks: ", tostring(err))
    end
    count = count and count or 0
    return count
end



function _M.incr_block_count()
    local _, err = shm_metrics:incr("block_count", 1, 0, 0)
    if err then
        log.error("Failed to increase the number of blocks: ", tostring(err))
    end
end

local qps = 0
do
    last_requests = 0
    function _M.calc_qps()
        total_requests = tonumber(C.ngx_stat_requests[0])
        incr_requests = total_requests - last_requests
        qps = math.floor(incr_requests / CALC_QPS_INTERVAL)
        last_requests = total_requests
    end
end


function _M.show()
    return {
        nginx_status = get_nginx_status(),
        shm_status   = get_shm_status(),
        block_count  = get_block_count(),
        qps = qps,
        -- TODO: collect all worker lua vm
        lua_vm = utils.format_capacity(collectgarbage("count") *1024),
        worker_id = ngx.worker.id()
    }
end


return _M