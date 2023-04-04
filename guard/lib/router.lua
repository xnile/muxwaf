local require   = require
local tree      = require("tree")
local table_new = table.new
local ngx       = ngx
local ipairs    = ipairs
local error     = error


local _M = {
    _VERSION = 0.1
}

local METHODS = {"GET", "POST", "DELETE", "PATCH", "PUT", "OPTIONS", "HEAD"}


local function set_ctx_param(ctx, param)
    ctx:set_param(param)
    return
end


local function start(self, ctx)
    local uri = ctx.var.uri
    local method = ctx.var.request_method

    if not self.mux[method] then
        return ctx.say_404()
    end

    local handler, param = self.mux[method]:get(uri)
    if not handler then
        return ctx.say_404()
    end

    set_ctx_param(ctx, param)

    handler(ctx)
end


function _M.new(prefix)
    local mux = table_new(0, #METHODS)
    for _, method in ipairs(METHODS) do
        mux[method] = tree.new()
    end

    local self = {
        mux = mux,
        start = start,
    }

    return setmetatable(self, {
        __index = function(self, key)
            return function(self, path, handler)
                if not self or not path or not handler then
                    error("missing parameters", 2)
                    return
                end

                if not path or path == '' then
                    error("path cannot be empty", 2)
                    return
                end

                if not handler then
                    error("handler cannot be empty", 2)
                    return
                end

                if not self.mux[key] then
                    error(key .. " method is not supported", 2)
                    return
                end

                self.mux[key]:insert(prefix .. path, handler)
            end
        end
    })
end

return _M