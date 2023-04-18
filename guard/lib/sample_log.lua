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
local assert         = assert
local io_open        = io.open
local io_close       = io.close
local table_clone    = require("table.clone")

local _M = {
  _VERSION = 0.1
}

local shm_log = ngx_shared[constants.DICTS.LOG]
local ACTION_TYPES = constants.ACTION_TYPES
local SAMPLE_LOG_FILE = ngx_cfg_prefix() .. "logs/sampled.log"
local sample_log_file_fd


local config = {
    is_sample_log_upload = 0,
    sample_log_upload_api = "",
    sample_log_upload_api_token = "",
}


local function upload_sample_log(log)
    -- local uri = "http://127.0.0.1:8001/api/attack-logs"
    local url = config.sample_log_upload_api
    if not url or url == "" then
        ngx_log(ngx.ERR, "failed to update sample log: ", "api url is empty.")
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
        ngx_log(ngx.ERR, "failed to update sample log: ", err)
    end
end

local function disable_upload_sample_log()
    config.is_sample_log_upload = 0
    config.sample_log_upload_api = ""
    config.sample_log_upload_api_token = ""
end

local function update_log_config(cfg)
    if not cfg.is_sample_log_upload or not cfg.sample_log_upload_api or not cfg.sample_log_upload_api_token then
        ngx_log(ngx.ERR, "failed to update log configuration: parameter error")
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

local function sampled(ctx, rule_type, action)
    local now = time.now()
    local raw_log = {
        content = {
            host               = ctx.var.host,
            site_id            = ctx.site_id,
            real_client_ip     = ctx.real_client_ip,
            request_id         = ctx.var.request_id,    
            remote_addr        = ctx.var.remote_addr,
            request_path       = ctx.var.request_uri,
            request_method     = ctx.var.request_method,
            request_time       = math.floor(ctx.var.msec),
            waf_start_time     = math.floor(ctx.waf_start_time /1000 /1000),
            waf_process_time   = now - ctx.waf_start_time,
            rule_type          = rule_type,
            action             = action or -1,
            ngx_worker_id      = ctx.worker_id,
        }
    }

    local json_log = ctx.encode(raw_log)
    if not json_log then
        ngx_log(ngx.WARN, "faild to encode sampled log")
        return
    end

    shm_log:rpush("sample", json_log)
end

function _M.block(ctx, rule_type)
    if not metrics then
        metrics = require("metrics")
    end
    metrics.incr_block_count(rule_type)
    sampled(ctx, rule_type, ACTION_TYPES.BLOCK)
end

function _M.bypass(ctx, rule_type)
    sampled(ctx, rule_type, ACTION_TYPES.BYPASS)
end

function _M.iterator()
  local len = shm_log:llen("sample")
  for i = 1, len do
    local log, err = shm_log:lpop("sample")
    if not log then
      ngx_log(ngx.WARN, "failed to pop sampled log from shm: ", err)
      goto continue
    end

    sample_log_file_fd:write(log)
    sample_log_file_fd:write("\r\n")
    sample_log_file_fd:flush()

    if config.is_sample_log_upload == 1 then
        upload_sample_log(log)
    end

  ::continue::
  end
end

function _M.worker_exit(self)
    io_close(sample_log_file_fd)
end


return _M