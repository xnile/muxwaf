local error = error
local ngx = ngx
local ngx_log = ngx.log
local ERR = ngx.ERR
local WARN= ngx.WARN
local setmetatable = setmetatable
local ffi = require("ffi")
local ffi_cast = ffi.cast
local ffi_load = ffi.load
local ffi_cdef = ffi.cdef
local ffi_string = ffi.string
local string_format = string.format
local table_new = table.new

local _M = {}

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


local lib_name = "resty/librestyradixtree.so"
if ffi.os == "OSX" then
    lib_name = "resty/librestyradixtree.dylib"
end


local radix, tried_paths = load_shared_lib(lib_name)
if not radix then
    tried_paths[#tried_paths + 1] = 'tried above paths but can not load '
                                    .. lib_name
    error(table.concat(tried_paths, '\r\n', 1, #tried_paths))
end



ffi_cdef[[
    void *radix_tree_new();
    int radix_tree_destroy(void *t);
    int radix_tree_insert(void *t, const unsigned char *buf, size_t len,
        void *cdata);
    int radix_tree_remove(void *t, const unsigned char *buf, size_t len);
    void *radix_tree_find(void *t, const unsigned char *buf, size_t len);
    void *radix_tree_search(void *t, void *it, const unsigned char *buf,
        size_t len);
    int radix_tree_prev(void *it, const unsigned char *buf, size_t len);
    int radix_tree_stop(void *it);
    void *radix_tree_new_it(void *t);
    void *radix_tree_get_data_by_it(void *it);
]]

local function gc_free(self)
    local it = self.tree_it
    if it then
        radix.radix_tree_stop(it)
        C.free(it)
        self.tree_it = nil
    end

    if self.tree then
        radix.radix_tree_destroy(self.tree)
        self.tree = nil
    end

    return
end

-- https://stackoverflow.com/questions/27426704/lua-5-1-workaround-for-gc-metamethod-for-tables
local function setmt__gc(t, mt)
  local prox = newproxy(true)
  getmetatable(prox).__gc = function() mt.__gc(t) end
  t[prox] = true
  return setmetatable(t, mt)
end

local _mt = { __index = _M, __gc = gc_free }

function _M.new()
  local trie = radix.radix_tree_new()
  local trie_it = radix.radix_tree_new_it(trie)
  if trie_it == nil then
    error("failed to new radix trie iterator")
  end

  local self = {
    trie = trie,
    trie_it  = trie_it
  }

  if _G._VERSION <= "Lua 5.1" then
    return setmt__gc(self, _mt)
  else
    return setmetatable(self, _mt)
  end
 
end

-- path = { id = "xxx", path = "/xxx/xxx/xxx"} 
function _M.insert(self, path, value)
  if type(path) ~= "string" then
    return false, "path should be a string"
  end

  if type(value) ~= "string" then
    return false, "value should be a string"
  end

  value = ffi_cast("char*",value)
  local result = radix.radix_tree_insert(self.trie, path, #path, value)
  -- 1: insert; 0: update;
  return result >= 0, string_format("insert path: %s and value: %s error", path, value)
end


function _M.remove(self, path)
  if type(path) ~= "string" then
    return false, "path should be a string"
  end

  -- Returns 1 if the item was found and deleted, 0 otherwise.
  local result = radix.radix_tree_remove(self.trie, path, #path)
  if result == 1 then
    return true, nil
  end
  return false, string_format("remove path: %s error",path)
end


function _M.prefix_find(self, path)
  if type(path) ~= "string" then
    return  nil, "path should be a string"
  end
  
  local it = radix.radix_tree_search(self.trie, self.trie_it, path, #path)
  if not it then
    return nil, "failed to prefix find"
  end


  -- Returns 1 if the item was found and deleted, -1 otherwise.
  local result = radix.radix_tree_prev(it, path, #path)
  if result == 1 then
    local cdata = radix.radix_tree_get_data_by_it(it)
    if cdata ~= nil then
      return ffi_string(cdata), nil
    else
      return nil, "failed to prefix find"
    end
  end
  return nil, nil
end

function _M.exact_find(self, path)
  if type(path) ~= "string" then
    return  nil, "path should be a string"
  end
  
  local cdata = radix.radix_tree_find(self.trie, path, #path)
  if cdata ~= nil then
    return ffi_string(cdata), nil
  end

  return nil, nil
end

return _M