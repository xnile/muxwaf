local ffi = require("ffi")

local _M = {
    _VERSION = 0.1
}
 
ffi.cdef[[
    struct timeval {
        long int tv_sec;
        long int tv_usec;
    };
    int gettimeofday(struct timeval *tv, void *tz);
]]
 
local tv = ffi.new("struct timeval")

-- in microseconds
function _M.now()
    ffi.C.gettimeofday(tv, nil)
    return tonumber(tv.tv_sec) * 1000000 + tonumber(tv.tv_usec)
end

return _M
