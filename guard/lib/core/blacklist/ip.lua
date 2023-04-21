local require        = require
local ipmatcher      = require("resty.ipmatcher")
local log            = require("log")
local ipairs         = ipairs
local setmetatable   = setmetatable
-- local string_format  = string.format
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
      log.warn("Failed to add the IP address \"",ip, "\" to the IP blacklist, the rule with ID \"",id, "\" already exists")
      goto continue
    end

    local ok, err = tree:insert(ip, id)
    if not ok then
      log.warn("Failed to add the IP address \"",ip, "\" to the IP blacklist, ", err)
      goto continue
    end

    cache[id] = table_clone(item)
    log.debug("successed to add the IP address \"", ip, "\" to the IP blacklist")

    ::continue::
  end
end


function _M.del(self, items)
  local cache, trie = cache, trie

  for _, id in ipairs(items) do
    local item = cache[id]

    if not item then
      log.warn("failed to remove IP blacklist, the rule with ID \"", id, "\" does not exist")
      goto continue
    end

    local ip = item.ip
    local ok, err = tree:remove(ip)
    
    if not ok then
      log.error("failed to remove the IP address \"", ip, "\" from the IP blacklist, ", err)
      goto continue
    end

    cache[id] = nil
    log.debug("successed to remove the IP address \"", ip, "\" from the IP blacklist")

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
      log.debug("successed to add the IP address \"", ip, "\" to the IP blacklist")
      new_cache[id] = table_clone(item)
    else
      log.error("failed to add the IP address \"", ip, "\" to the IP blacklist") 
    end
  end
  cache = new_cache
  tree = new_tree
end


-- @tparam string ip
-- @return boolen, string [,string]
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