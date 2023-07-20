local net = require("utils.net")
local ffi = require("ffi")
local ffi_new = ffi.new
local C       = ffi.C
local table_new = table.new
local ngx_re = require "ngx.re"
local setmetatable = setmetatable

ffi.cdef[[
    void *xdb_new(const char *db_path);
    int xdb_find(void *xdb, const char *str_ip, char *buf, size_t len);
    int xdb_destroy(void *xdb);
]]


local function load_shared_lib(so_name)
    local string_gmatch = string.gmatch
    local string_match = string.match
    local io_open = io.open
    local io_close = io.close

    local cpath = package.cpath
    local tried_paths = table_new(32, 0)
    local i = 1

    for k, _ in string_gmatch(cpath, "[^;]+") do
        local fpath = string_match(k, "(.*/)")
        fpath = fpath .. so_name
        local f = io_open(fpath)
        if f ~= nil then
            io_close(f)
            return ffi.load(fpath)
        end
        tried_paths[i] = fpath
        i = i + 1
    end

    return nil, tried_paths
end

local lib_name = "geo/parser/ip2region/libxdb.so"
if ffi.os == "OSX" then
    lib_name = "geo/parser/ip2region/libxdb.dylib"
end

local xdb, tried_paths = load_shared_lib(lib_name)
if not xdb then
    tried_paths[#tried_paths + 1] = 'tried above paths but can not load '
                                    .. lib_name
    error(table.concat(tried_paths, '\r\n', 1, #tried_paths))
end


local _M = {
    _VERSION = 0.1
}

local IP_DATA_BUF_MAX_SIZE = 255
local str_buf
local c_buf_type = ffi.typeof("char[?]")

local function gc_free(self)
    local xdb_cdata = self.xdb
    if xdb_cdata ~= nil then
        xdb.xdb_destroy(xdb_cdata)
        C.free(xdb_cdata)
        self.xdb = nil
    end
end

-- https://stackoverflow.com/questions/27426704/lua-5-1-workaround-for-gc-metamethod-for-tables
local function setmt__gc(t, mt)
  local prox = newproxy(true)
  getmetatable(prox).__gc = function() mt.__gc(t) end
  t[prox] = true
  return setmetatable(t, mt)
end

local _mt = { __index = _M, __gc = gc_free }


-- local xdb_path = "/Users/xnile/xnile/workspace/mygit/code-reading/ip2region-master/world.xdb"
function _M.new(xdb_path)
    local xdb_cdata = xdb.xdb_new(xdb_path)

    -- cannot use `not xdb_cdata`
    if xdb_cdata == nil then
        error("failed to new xdb searcher", 2)
    end

    local self = {
        xdb = xdb_cdata
    }

    if _G._VERSION <= "Lua 5.1" then
        return setmt__gc(self, _mt)
    else
        return setmetatable(self, _mt)
    end
end


-- xdb 格式的返回值格式：州、国家、省份、城市、区县、一级行政代码、二级行政代码、三级行政代码、国家英文、国家代码、国际区号、运营商、lat、lng
local XDB_FIELDS = {"continent", "country_name", "region_name", "city_name", "district_name", "china_admin_code1", "china_admin_code2", "china_admin_code3", "country", "country_code", "idd_code", "isp", "lat", "lng"}
local function parse_location(raw)
    local res, err = ngx_re.split(raw, '\\|')
    if not res then
        return nil, err
    end

    local data = table_new(0, 14)
    for k, v in ipairs(res) do
        data[XDB_FIELDS[k]] = v
    end
    return data, nil
end


function _M.search(self, ip)
    -- if type(ip) ~= "string" then
    --     return nil, "ip should a string"
    -- end

    -- if is_valid_ipv6(ip) then
    --     return nil, "IPv6 is not currently supported"
    -- end

    -- if not net.is_valid_ip(ip) then
    --     return nil, "'" .. ip .. "' is an invalid ipv4"
    -- end

    if not str_buf then
        str_buf = ffi_new(c_buf_type, IP_DATA_BUF_MAX_SIZE)
    end

    local rc = xdb.xdb_find(self.xdb, ip, str_buf, IP_DATA_BUF_MAX_SIZE)
    if rc == 0 then
        return parse_location(ffi.string(str_buf))
    end

    return nil, "Failed to search for '" .. ip .. "'"
end

return _M