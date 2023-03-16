local dns         = require("utils.net.dns")
local ffi         = require "ffi"
local C           = ffi.C
local ffi_cdef    = ffi.cdef
local ffi_new     = ffi.new


local AF_INET     = 2
local AF_INET6    = 10

if ffi.os == "OSX" then
    AF_INET6 = 30
end

    
local _M = {
    dns = dns,
    _VERSION = 0.1
}


ffi_cdef[[
    int inet_pton(int af, const char * restrict src, void * restrict dst);
]]


function _M.is_valid_ip(ip)
    if type(ip) ~= "string" then
        return false
    end

    local inet = ffi_new("unsigned int [1]")
    if C.inet_pton(AF_INET, ip, inet) == 1 then
        return true
    end

    if C.inet_pton(AF_INET6, ip, inet) == 1 then
        return true
    end

    return false

end

return _M