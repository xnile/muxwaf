local require     = require
local constants   = require("constants")
local utils       = require("utils")
local log         = require("log")
local ffi         = require "ffi"
local C           = ffi.C
local tonumber    = tonumber
local tostring    = tostring
local math        = math
local ipairs      = ipairs
local table_new   = table.new
local ngx_shared  = ngx.shared

local _M = {
    _VERSION = 0.1
}


local CALC_QPS_INTERVAL = constants.CALC_QPS_INTERVAL
local RULE_TYPE         = constants.RULE_TYPE
local BLOCK_RULE_TYPE   = { RULE_TYPE.BLACKLIST_IP, RULE_TYPE.BLACKLIST_REGION, RULE_TYPE.RATELIMIT }

local ditcs       = constants.DICTS
local shm_metrics = ngx_shared[ditcs.METRICS]
local resp_sts_code_shm_keys


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


-- local function get_block_count()

--     local count, err = shm_metrics:get("block_count")
--     if err then
--         log.error("Failed to get the number of blocks: ", tostring(err))
--     end
--     count = count and count or 0
--     return count
-- end

local function get_block_count()
    local block_count = {}
    local total = 0
    for _, rule_type in ipairs(BLOCK_RULE_TYPE) do
        local count, err = shm_metrics:get("block_count_" .. rule_type)
        if err then
            log.error("failed to get the number of blocks: ", tostring(err))
        end
        count = count and count or 0
        block_count[rule_type] = count
        total = total + count
    end
    block_count['total'] = total
    return block_count
end


-- function _M.incr_block_count()
--     local _, err = shm_metrics:incr("block_count", 1, 0, 0)
--     if err then
--         log.error("Failed to increase the number of blocks: ", tostring(err))
--     end
-- end


function _M.incr_block_count(rule_type)
    local _, err = shm_metrics:incr("block_count_" .. rule_type, 1, 0, 0)
    if err then
        log.error("failed to increase the number of blocks: ", tostring(err))
    end
end


local qps = 0
do
    last_requests = 0
    function _M.calc_qps()
        local total_requests = tonumber(C.ngx_stat_requests[0])
        local incr_requests = total_requests - last_requests
        qps = math.floor(incr_requests / CALC_QPS_INTERVAL)
        last_requests = total_requests
    end
end


-- TODO: better it
function _M.incr_resp_sts_code()
    local code = ngx.var.status
    local shm_key = ''
    if code == '200' then
        shm_key = '200'
    elseif code == '204' then
        shm_key = '204'
    elseif code == '403' then
        shm_key = '403'
    elseif code == '404' then
        shm_key = '404'
    elseif code == '499' then
        shm_key = '499'
    elseif code == '500' then
        shm_key = '500'
    elseif code  == '502' then
        shm_key = '502'
    elseif code == '504' then
        shm_key = '504'
    else
        shm_key = 'other_no_200'
    end

    local _, err = shm_metrics:incr("resp_sts_code_" .. shm_key, 1, 0, 0)
    if err then
        log.error("failed to increase response status code: ", tostring(err))
    end
end


resp_sts_code_shm_keys = { '200', '204', '403' ,'404', '499', '500', '502', '504', 'other_no_200'}
local function get_resp_sts_code_count()
    local r = {}
    for _, key in ipairs(resp_sts_code_shm_keys)do
        local c, err = shm_metrics:get("resp_sts_code_" .. key)
        if err then
            log.error("failed to get the number of response status code: ", tostring(err))
        end
        r[key] = c and c or 0
    end
    return r
end


function _M.show()
    return {
        nginx_status = get_nginx_status(),
        shm_status   = get_shm_status(),
        block_count  = get_block_count(),
        qps = qps,
        -- TODO: collect all worker lua vm
        lua_vm = utils.format_capacity(collectgarbage("count") *1024),
        worker_id = ngx.worker.id(),
        rsp_sts_code = get_resp_sts_code_count(),
    }
end


return _M