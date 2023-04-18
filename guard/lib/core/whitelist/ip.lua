local require        = require
local ipmatcher      = require("resty.ipmatcher")
local log            = require("log")
local ipairs         = ipairs
local setmetatable   = setmetatable
local string_format  = string.format
local table_clear    = table.clear
local table_new      = table.new
local table_clone    = require("table.clone")

local cache = table_new(0, 100)
local tree  = ipmatcher.new()

local _M = {}

function _M.add(self, items)
  for _, item in ipairs(items) do
    local id, ip = item.id, item.ip

    if cache[id] then
      log.warn(string_format("faild add whitelist ip '%s': the rule id '%s' already exists",ip, id))
      goto continue
    end

    local ok, err = tree:insert(ip, id)
    if not ok then
      log.warn(string_format("faild add whitelist ip '%s': '%s'",ip, err))
      goto continue
    end

    cache[id] = table_clone(item)
    log.debug(string_format("add whitelist ip '%s' success", ip))

    ::continue::
  end
end


function _M.del(self, items)
  local cache, trie = cache, trie

  for _, id in ipairs(items) do
    local item = cache[id]

    if not item then
      log.warn(string_format("faild add whitelist ip, the rule id '%s' does not exist", id))
      goto continue
    end

    local ok, err = tree:remove(item.ip)
    
    if not ok then
      log.error(string_format("faild delete whitelist ip %s: %s", ip, err))
      goto continue
    end

    cache[id] = nil
    log.debug("delete whitelist ip '%s' success", ip)

    ::continue::
  end
  return
end

function _M.full_sync(_, items)
  local new_cache  = {}
  local new_tree = ipmatcher.new()

  for _, item in ipairs(items) do
    local id, ip = item.id, item.ip
    local ok, err = new_tree:insert(ip, id)
    if ok then
      new_cache[id] = table_clone(item)
    else
      log.error(string_format("failed full sync whitelist ip '%s': %s", ip, err))
    end
  end
  cache = new_cache
  tree = new_tree
  log.debug("full sync whitelist ip success")
end


-- tparam string ip
-- treturn boolean, number [, string]
function _M.match(self, ip)
  return tree:match(ip)
end


function _M.reset(self)
  table_clear(cache)
  tree = ipmatcher.new()
end


function _M.get_raw(_)
  local cnt = {}
  for _, item in pairs(cache) do
    cnt[#cnt +1] = table_clone(item)
  end
  return cnt
end

return _M