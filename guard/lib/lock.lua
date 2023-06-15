local lrucache      = require("resty.lrucache")

local _M = {
    _VERSION = 0.1
}


function _M.new(size)
    local c, err = lrucache.new(100)
    if not c then
        return nil, "failed to create the cache: " .. (err or "unknown")
    end

    local self = {
        cache = c
    }

    return setmetatable(self, { __index = _M}), nil
end


function _M.lock(self, key, ttl, flags)
    if not self.cache:get(key) then
        self.cache:set(key, true, ttl)
        return true
    end
    return false
end


function _M.unlock(self, key)
    self.cache:delete(key)
end


return _M