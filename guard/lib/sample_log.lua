local log            = require("log")
local time           = require("time")
local metrics        = require("metrics")
local constants      = require("constants")
local http           = require("resty.http")
local ngx_errlog     = require("ngx.errlog")
local ngx            = ngx
local ngx_log        = ngx.log
local ngx_shared     = ngx.shared
local ngx_get_phase  = ngx.get_phase
local ngx_cfg_prefix = ngx.config.prefix
local ngx_re_sub     = ngx.re.sub
local assert         = assert
local io_open        = io.open
local io_close       = io.close
local table_new      = table.new
local table_clear    = table.clear
local table_concat   = table.concat
local table_clone    = require("table.clone")

local _M = {
  _VERSION = 0.1
}

-- The number of log entries sent in each batch
local BATCH_SIZE = 10000
local ACTION_TYPES = constants.ACTION_TYPES
local SAMPLE_LOG_FILE = ngx_cfg_prefix() .. "logs/sampled.log"

local shm_log = ngx_shared[constants.DICTS.LOG]
local sample_log_batch = table_new(BATCH_SIZE +2, 0)
local sample_log_file_fd


local config = {
    is_sample_log_upload = 0,
    sample_log_upload_api = "",
    sample_log_upload_api_token = "",
}


local function send_sample_log(log)
    local url = config.sample_log_upload_api
    if not url or url == "" then
        ngx_log(ngx.ERR, "failed to post sample log: ", "api url is empty")
        return
    end

    local httpc = http.new()
    local res, err = httpc:request_uri(url,{
        method = "POST",
        body = log,
        headers = {
            ["Content-Type"] = "application/json",
            ["Token"] = config.sample_log_upload_api_token,
        },
    })

    if not res then
        ngx_log(ngx.ERR, "failed to post sample log: ", err)
    end
end

local function disable_upload_sample_log()
    config.is_sample_log_upload = 0
    config.sample_log_upload_api = ""
    config.sample_log_upload_api_token = ""
end

local function update_log_config(cfg)
    if not cfg.is_sample_log_upload or not cfg.sample_log_upload_api or not cfg.sample_log_upload_api_token then
        ngx_log(ngx.ERR, "failed to update sample log configuration: parameter error")
        return
    end
    config = table_clone(cfg)
end



function _M.init()
    local fd = io_open(SAMPLE_LOG_FILE, "a")
    assert(fd, SAMPLE_LOG_FILE.. " sampled log file open failed")
    sample_log_file_fd = fd
end

-- update alias for uniform
function _M.add(_, cfg)
    update_log_config(cfg)
end

function _M.update(_, cfg)
    update_log_config(cfg)
end

function _M.full_sync(_, cfg)
    update_log_config(cfg)
    ngx_log(ngx.DEBUG, "full sync log configuration success")
end

function _M.reset(_)
    disable_upload_sample_log()
end

function _M.get_raw(_)
    return table_clone(config)
end

local function sampled(ctx, rule_type, action, rule_id)
    ctx.sample_log =  {
        host               = ctx.var.host,
        site_id            = ctx.site_id,
        real_client_ip     = ctx.real_client_ip,
        request_id         = ctx.var.request_id,    
        remote_addr        = ctx.var.remote_addr,
        request_path       = ctx.var.request_uri,
        request_method     = ctx.var.request_method,
        request_time       = math.floor(ctx.var.msec),
        waf_start_time     = math.floor(ctx.waf_start_time /1000 /1000),
        waf_process_time   = time.now() - ctx.waf_start_time,
        rule_type          = rule_type,
        action             = action or -1,
        ngx_worker_id      = ctx.worker_id,
        rule_id            = rule_id,
        location           = ctx.location,
        time_local         = ctx.var.time_local,
    }
end

function _M.block(ctx, rule_type, rule_id)
    if not metrics then
        metrics = require("metrics")
    end
    metrics.incr_block_count(rule_type)
    sampled(ctx, rule_type, ACTION_TYPES.BLOCK, rule_id)
end

function _M.bypass(ctx, rule_type, rule_id)
    -- the bypass actions are not logged, TODO: configurable
    -- #sampled(ctx, rule_type, ACTION_TYPES.BYPASS, rule_id)
end


-- TODO: make it better
local iterator_running = false
function _M.iterator()
    if iterator_running then return end

    iterator_running = true
    local len = shm_log:llen("sample")
    -- ngx_log(ngx.DEBUG, "have ", tostring(len), "sample log")

    if #sample_log_batch == 0 then
        sample_log_batch[1] = "["
    end
    for i = 1, len do
        local log, err = shm_log:lpop("sample")
        if not log then
            ngx_log(ngx.WARN, "failed to pop sample log from shm: ", err)
            goto continue
        end

        local len = #sample_log_batch
        sample_log_batch[len+1] = log .. ","

        if len + 2 == BATCH_SIZE + 2 then
            sample_log_batch[len+1] = log
            sample_log_batch[len+2] = "]"
            local payload = table_concat(sample_log_batch)
            send_sample_log(payload)
            table_clear(sample_log_batch)
            sample_log_batch[1] = "["
        end

        ::continue::
    end

    if #sample_log_batch > 1 then
        local last = ngx_re_sub(sample_log_batch[#sample_log_batch], "(.*),$", "$1")
        sample_log_batch[#sample_log_batch] = last
        sample_log_batch[#sample_log_batch+1] = "]"
        local payload = table_concat(sample_log_batch)
        send_sample_log(payload)
        table_clear(sample_log_batch)
    end

    iterator_running = false
end

function _M.log_phase(ctx)
    if not ctx.sample_log or not next(ctx.sample_log) then return end

    sample_log_file_fd:write(ctx.encode(ctx.sample_log))
    sample_log_file_fd:write("\r\n")
    -- sample_log_file_fd:flush()

    local raw_log = {
        content = ctx.sample_log
    }
    if config.is_sample_log_upload == 1 then
        if shm_log:free_space() < 2 *1024*1024 then
            ngx_log(ngx.WARN, "omitting sample log, as the shm-based dictionary free page size is not enough.")
            return
        end

        local _, err = shm_log:rpush("sample", ctx.encode(raw_log))
        if err then
            ngx_log(ngx.ERR,"failed to push sample log to shm: ", err)
        end
    end
end

function _M.worker_exit(self)
    sample_log_file_fd:flush()
    io_close(sample_log_file_fd)
end


return _M