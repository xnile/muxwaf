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

local CALC_QPS_INTERVAL       = constants.CALC_QPS_INTERVAL
local CALC_BANDWIDTH_INTERVAL = constants.CALC_BANDWIDTH_INTERVAL
local RULE_TYPE         = constants.RULE_TYPE
local BLOCK_RULE_TYPE   = { RULE_TYPE.BLACKLIST_IP, RULE_TYPE.BLACKLIST_REGION, RULE_TYPE.RATELIMIT }

-- response status codes that need to be counted
local RSP_STS_CODE_COUNTED      = { '200', '400', '403' ,'404', '413', '499', '500', '502', '504', 'other' }
local UPSTREAM_STS_CODE_COUNTED = { '200', '400', '403' ,'404', '413', '499', '500', '502', '504', 'other' }

-- shm keys
local KEY_QPS                      = "qps"
local KEY_LAST_TOTAL_REQUESTS      = "last_total_requests"
local KEY_BLOCK_COUNT_PREFIX       = "block_count_"
local KEY_RSP_STS_CODE_PREFIX      = "rsp_sts_code_"
local KEY_UPSTREAM_STS_CODE_PREFIX = "upstream_sts_code_"
local KEY_TRAFFIC_IN               = "traffic_in"
local KEY_TRAFFIC_OUT              = "traffic_out"
local KEY_LAST_TRAFFIC_IN          = "last_trffic_in"
local KEY_LAST_TRAFFIC_OUT         = "last_trffic_out"
local KEY_BANDWIDTH_IN             = "bandwidth_in"
local KEY_BANDWIDTH_OUT            = "bandwidth_out"


local DICTS       = constants.DICTS
local shm_metrics = ngx_shared[DICTS.METRICS]


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

local function get_shm_status(pretty)
    local status = table_new(0, 3)
    for _, dict in pairs(DICTS) do
        local shm = ngx_shared[dict]

        if pretty then
            status[dict] = {
                capacity = utils.pretty_bytes(shm:capacity()),
                free     = utils.pretty_bytes(shm:free_space()),
            }
        else
            status[dict] = {
                capacity = shm:capacity(),
                free     = shm:free_space(),
            }
        end            
    end
    return status
end


local function get_block_count()
    local block_count = {}
    local total = 0
    for _, rule_type in ipairs(BLOCK_RULE_TYPE) do
        local count = shm_metrics:get("block_count_" .. rule_type)
        count = count or 0
        block_count[rule_type] = count
        total = total + count
    end
    block_count['total'] = total
    return block_count
end


function _M.incr_block_count(rule_type)
    local _, err = shm_metrics:incr("block_count_" .. rule_type, 1, 0, 0)
    if err then
        log.error("failed to increase the number of blocks: ", tostring(err))
    end
end



function _M.calc_qps()
    local last_total_requests = shm_metrics:get("KEY_LAST_TOTAL_REQUESTS")
    last_total_requests = last_total_requests or 0

    local total_requests = tonumber(C.ngx_stat_requests[0])
    local incr_requests = total_requests - last_total_requests
    local qps = math.floor(incr_requests / CALC_QPS_INTERVAL + 0.5)

    local ok, err = shm_metrics:set(KEY_QPS, qps)
    if not ok then
        log.error("failed to set qps: ", tostring(err))
    end

    local ok, err = shm_metrics:set("last_total_requests", total_requests)
    if not ok then
        log.error("failed to set last total requests: ", tostring(err))
    end
end


local function get_qps()
    local qps = shm_metrics:get(KEY_QPS)
    return qps or 0
end


function _M.incr_resp_sts_code()
    local code = ngx.var.status
    if code == '204' then
        return  -- ignore 204 status code
    end

    -- local shm_key = 'other'
    local shm_key = RSP_STS_CODE_COUNTED[#RSP_STS_CODE_COUNTED]

    for _, c in ipairs(RSP_STS_CODE_COUNTED) do
        if code == c then
            shm_key = code
        end
    end

    local _, err = shm_metrics:incr(KEY_RSP_STS_CODE_PREFIX .. shm_key, 1, 0, 0)
    if err then
        log.error("failed to increase response status code: ", tostring(err))
    end
end

local function get_resp_sts_code_count()
    local r = {}
    for _, key in ipairs(RSP_STS_CODE_COUNTED)do
        local c = shm_metrics:get(KEY_RSP_STS_CODE_PREFIX .. key)
        r[key] = c or 0
    end
    return r
end


function _M.incr_upstream_sts_code()
    local code = ngx.var.upstream_status
    if not code then return end

    if code == '204' then
        return  -- ignore 204 status code
    end

    local shm_key = UPSTREAM_STS_CODE_COUNTED[#UPSTREAM_STS_CODE_COUNTED]

    for _, c in ipairs(UPSTREAM_STS_CODE_COUNTED) do
        if code == c then
            shm_key = code
        end
    end

    local _, err = shm_metrics:incr(KEY_UPSTREAM_STS_CODE_PREFIX .. shm_key, 1, 0, 0)
    if err then
        log.error("failed to increase response status code: ", tostring(err))
    end
end

local function get_upstream_sts_code_count()
    local r = {}
    for _, key in ipairs(UPSTREAM_STS_CODE_COUNTED)do
        local c = shm_metrics:get(KEY_UPSTREAM_STS_CODE_PREFIX .. key)
        r[key] = c or 0
    end
    return r
end


function _M.incr_traffic()
    local i = tonumber(ngx.var.request_length) or 0
    local o = tonumber(ngx.var.bytes_sent) or 0

    do
        local _, err = shm_metrics:incr(KEY_TRAFFIC_IN, i, 0, 0)
        if err then
            log.error("failed to increase traffic in: ", tostring(err))
        end
    end

    do
        local _, err = shm_metrics:incr(KEY_TRAFFIC_OUT, o, 0, 0)
        if err then
            log.error("failed to increase traffic out: ", tostring(err))
        end
    end
end

local function get_traffic(pretty)
    local into = shm_metrics:get(KEY_TRAFFIC_IN) or 0
    local out  = shm_metrics:get(KEY_TRAFFIC_OUT) or 0

    if pretty then
        into = utils.pretty_bytes(into)
        out  = utils.pretty_bytes(out)
    end

    return {
        ["in"] = into,
        out    = out
    }
end

function _M.calc_bandwidth()
    local traffic_in       = shm_metrics:get(KEY_TRAFFIC_IN) or 0
    local traffic_out      = shm_metrics:get(KEY_TRAFFIC_OUT) or 0
    local last_traffic_in  = shm_metrics:get(KEY_LAST_TRAFFIC_IN) or 0
    local last_traffic_out = shm_metrics:get(KEY_LAST_TRAFFIC_OUT) or 0


    local bandwidth_in  = (traffic_in - last_traffic_in) / CALC_BANDWIDTH_INTERVAL *8
    local bandwidth_out = (traffic_out - last_traffic_out) / CALC_BANDWIDTH_INTERVAL *8

    do
        local ok, err = shm_metrics:set(KEY_BANDWIDTH_IN, bandwidth_in)
        if not ok then
            log.error("failed to set bandwidth in: ", tostring(err))
        end
    end

    do
        local ok, err = shm_metrics:set(KEY_BANDWIDTH_OUT, bandwidth_out)
        if not ok then
            log.error("failed to set bandwidth out: ", tostring(err))
        end
    end


    do
        local ok, err = shm_metrics:set(KEY_LAST_TRAFFIC_IN, traffic_in)
        if not ok then
            log.error("failed to set last traffic in: ", tostring(err))
        end
    end

    do
        local ok, err = shm_metrics:set(KEY_LAST_TRAFFIC_OUT, traffic_out)
        if not ok then
            log.error("failed to set last traffic out: ", tostring(err))
        end
    end

end


local  function get_bandwidth()
    local bandwidth_in  = shm_metrics:get(KEY_BANDWIDTH_IN) or 0
    local bandwidth_out = shm_metrics:get(KEY_BANDWIDTH_OUT) or 0
    return {
        ["in"]  = bandwidth_in,
        ["out"] = bandwidth_out,
    }
end


-- TODO: collect all worker lua vm
local function get_lua_vm(pretty)
    local lua_vm = collectgarbage("count") *1024
    if pretty then
        return utils.pretty_bytes(lua_vm)
    else
        return lua_vm
    end
end



function _M.show(ctx)
    local args, err = ngx.req.get_uri_args()
    if err then
        log.error("failed to get request args: ", tostring(err))
    end
    local pretty = args["pretty"] or false

    return {
        nginx_status      = get_nginx_status(),
        shm_status        = get_shm_status(pretty),
        block             = get_block_count(),
        qps               = get_qps(),
        lua_vm            = get_lua_vm(pretty),
        worker_id         = ngx.worker.id(),
        rsp_sts_code      = get_resp_sts_code_count(),
        upstream_sts_code = get_upstream_sts_code_count(),
        traffic           = get_traffic(pretty),
        bandwidth         = get_bandwidth(),
    }
end

return _M