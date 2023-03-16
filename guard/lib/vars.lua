local require   = require
local tablex    = require("utils.tablex")
local ngx_var   = ngx.var
local ipairs    = ipairs
local string_format = string.format


local NGX_VARS = {
    "uri",
    "host",
    "https",
    "scheme",
    "remote_addr",
    "realip_remote_addr",
    "request_id",
    "request_uri",
    "request_time",
    "request_method",
    "ssl_server_name",
    "server_port",
}

local CUSTOME_VARS = {
    "upstream_x_real_ip",
    "upstream_scheme"
}

local function getter(self, key)

    if type(key) ~= "string" then
        return error("invalid argument, string expect", 2)
    end

    local vars = tablex.array_merge(NGX_VARS, CUSTOME_VARS)
    if not tablex.array_contains(vars, key) then
        return error(string_format("var '%s' is not allowed to access", key), 2)
    end

    -- lazy caching
    -- self[key] = ngx_var[key]

    return ngx_var[key]
end


local function setter(_, key, value)
    if type(key) ~= "string" then
        return error("invalid argument, string expect", 2)
    end

    if not tablex.array_contains(CUSTOME_VARS, key) then
        return error(string_format("update var '%s' is not supported", key), 2)
    end
    ngx_var[key] = value
end

return setmetatable({}, {
    __index = getter,
    __newindex = setter
})