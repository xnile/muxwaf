local ngx_errlog    = require("ngx.errlog")
local ngx           = ngx
local ngx_log       = ngx.log
local string_format = string.format
local ngx_get_phase = ngx.get_phase


local _M = {
  _VERSION = 0.1
}

local ngx_log_levels = {
    stderr = ngx.STDERR,
    emerg  = ngx.EMERG,
    alert  = ngx.ALERT,
    crit   = ngx.CRIT,
    error  = ngx.ERR,
    warn   = ngx.WARN,
    notice = ngx.NOTICE,
    info   = ngx.INFO,
    debug  = ngx.DEBUG,
}


local function get_cnt_log_level()
    local cnt
    -- https://github.com/openresty/lua-nginx-module/issues/467#issuecomment-82647228
    if ngx_get_phase() ~= "init" then
        cnt = ngx.config.subsystem == "http" and ngx_errlog.get_sys_filter_level()
    end
    return cnt
end


setmetatable(_M, {
    __index = function(self, level)
        local log_level = ngx_log_levels[level]
        local cnt_level = get_cnt_log_level()
        local cmd
        if not log_level then
            ngx_log(ngx_log_levels.error, string_format("command '%s' is not supported", level))
            cmd = function() end
        elseif cnt_level and (log_level > cnt_level) then
            cmd = function() end
        else
            cmd = function(...)
                return ngx_log(log_level, ...)
            end
        end

        if ngx_get_phase() ~= "init" then
            self[level] = cmd
        end

        return cmd
    end
})

return _M