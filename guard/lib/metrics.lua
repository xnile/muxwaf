local constants       = require("constants")
local log             = require("log")
local ffi             = require "ffi"
local C               = ffi.C
local tonumber        = tonumber
local tostring        = tostring
local ngx_shared      = ngx.shared
local ngx_timer_every = ngx.timer.every
local ngx_worker_id   = ngx.worker.id

local _M = {
    _VERSION = 0.1
}

ffi.cdef[[
    uint64_t *ngx_stat_active;
    uint64_t *ngx_stat_reading;
    uint64_t *ngx_stat_writing;
    uint64_t *ngx_stat_waiting;
    /*
    uint64_t *ngx_stat_requests;
    uint64_t *ngx_stat_accepted;
    uint64_t *ngx_stat_handled;
    */
]]

local DICTS       = constants.DICTS
local shm_metrics = ngx_shared[DICTS.METRICS]
local DEFAULT_API_LISTEN_PORT = constants.DEFAULT_API_LISTEN_PORT

local prometheus

-- prometheus metrics definition
local prom_metric_requests
local prom_metric_requests_total
local prom_metric_request_duration_seconds
local prom_metric_connections
-- local prom_metric_connections_total
local prom_metric_response_bytes_total
local prom_metric_request_bytes_total
local prom_metric_attacks_blocked
local prom_metric_lua_memstats_alloc_bytes
local prom_metric_shm_total_bytes
local prom_metric_shm_free_bytes


local function collect_shm_mem_alloc()
    for _, dict in pairs(DICTS) do
        local shm = ngx_shared[dict]
        prom_metric_shm_total_bytes:set(shm:capacity(), {dict})
        prom_metric_shm_free_bytes:set(shm:free_space(), {dict})

    end
end


local function collect_lua_mem_alloc()
    local worker_id = ngx_worker_id()
    prom_metric_lua_memstats_alloc_bytes:set(collectgarbage("count") *1024, {worker_id})
end


function _M.init_worker()
    prometheus = require("resty.prometheus").init("muxwaf_metrics")

    prom_metric_requests = prometheus:counter("muxwaf_requests", "The total number of client requests", {"host", "status", "upstream_status"})
    prom_metric_requests_total = prometheus:counter("muxwaf_requests_total", "The total number of client requests")
    prom_metric_request_duration_seconds = prometheus:histogram("muxwaf_request_duration_second", "The request processing time in milliseconds", {"host"})
    prom_metric_connections = prometheus:gauge("muxwaf_connections", "current number of client connections with state {active, reading, writing, waiting}", {"state"})
    prom_metric_request_bytes_total = prometheus:counter("muxwaf_request_bytes_total", "total number of bytes request")
    prom_metric_response_bytes_total = prometheus:counter("muxwaf_response_bytes_total", "total number of bytes response")
    prom_metric_attacks_blocked = prometheus:counter("muxwaf_attacks_blocked", "number of attacks blocked", {"rule_type"})
    prom_metric_lua_memstats_alloc_bytes = prometheus:gauge("muxwaf_lua_memstats_alloc_bytes", "Number of bytes allocated and still in use", {"worker_id"})
    prom_metric_shm_total_bytes = prometheus:gauge("muxwaf_shm_total_bytes", "The shm-based dictionary capacity", {"name"})
    prom_metric_shm_free_bytes = prometheus:gauge("muxwaf_shm_free_bytes", "The shm-based dictionary free bytes", {"name"})

    local worker_id  = ngx_worker_id()
    do
        log.debug("start timer for collect lua memory allocated on worker ", worker_id)
        local ok, err = ngx_timer_every(15, collect_lua_mem_alloc)
        assert(ok, "failed to setting up timer for collect lua memory allocated: " .. tostring(err))
    end

    do 
        if worker_id == 0 then 
            log.debug("start timer for collect shm memory allocated on worker ", worker_id)
            local ok, err = ngx_timer_every(15, collect_shm_mem_alloc)
            assert(ok, "failed to setting up timer for collect shm memory allocated: " .. tostring(err))
        end
    end    
end


function _M.log_phase(_)
    -- ignore self
    if ngx.var.server_port == tostring(DEFAULT_API_LISTEN_PORT) and ngx.var.uri == "/api/sys/metrics" then return end

    local host            = ngx.var.host or "-"
    local status          = ngx.var.status or "-"
    local upstream_status = ngx.var.upstream_status or "-"
    local request_time    = tonumber(ngx.var.request_time) or -1
    local request_length  = tonumber(ngx.var.request_length) or -1
    local response_length = tonumber(ngx.var.bytes_sent) or -1

    prom_metric_requests:inc(1, {host, status, upstream_status})
    prom_metric_requests_total:inc(1)
    prom_metric_request_duration_seconds:observe(request_time, {host})
    prom_metric_connections:set(tonumber(C.ngx_stat_active[0]), {"active"})
    prom_metric_connections:set(tonumber(C.ngx_stat_reading[0]), {"reading"})
    prom_metric_connections:set(tonumber(C.ngx_stat_writing[0]), {"writing"})
    prom_metric_connections:set(tonumber(C.ngx_stat_waiting[0]), {"waiting"})
    prom_metric_request_bytes_total:inc(request_length)
    prom_metric_response_bytes_total:inc(response_length)
end


function _M.incr_block_count(rule_type)
    prom_metric_attacks_blocked:inc(1, {rule_type})
end

-- function _M.collect_lua_mem_alloc_timer_callback()
--     local worker_id = ngx_worker_id()
--     prom_metric_lua_memstats_alloc_bytes:set(collectgarbage("count") *1024, {worker_id})
-- end


function _M.collect(_)
    return prometheus:collect()
end

return _M