local require        = require
local ipmatcher      = require("resty.ipmatcher")
local log            = require("log")
local ipairs         = ipairs
local setmetatable   = setmetatable
local table_clear    = table.clear
local table_new      = table.new
local table_clone    = require("table.clone")

local raw = table_new(0, 100)
local tree  = ipmatcher.new()

local _M = {}

function _M.add(self, items)
  for _, item in ipairs(items) do
    local id, ip = item.id, item.ip

    if raw[id] then
      log.warn("The IP address '", ip, "' failed to be added to the IP blacklist, ", "the rule with ID '", id, "' already exists")
      goto continue
    end

    local ok, err = tree:insert(ip, id)
    if not ok then
      log.error("The IP address '", ip, "' failed to be added to the IP blacklist, ", err)
      goto continue
    end

    raw[id] = table_clone(item)
    log.debug("The IP address '", ip, "' has been successfully added to the IP blacklist")

    ::continue::
  end
end


function _M.del(self, items)
  local raw, trie = raw, trie

  for _, id in ipairs(items) do
    local item = raw[id]

    if not item then
      log.warn("failed to remove IP blacklist, the rule with ID '", id, "' does not exist")
      goto continue
    end

    local ip = item.ip
    local ok, err = tree:remove(ip)
    
    if not ok then
      log.error("The IP address '", ip, "' failed to be removed from the IP blacklist, ", err)
      goto continue
    end

    raw[id] = nil
    log.debug("The IP address '", ip, "' has been successfully removed from the IP blacklist")

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
      log.debug("The IP address '", ip, "' has been successfully added to the IP blacklist")
      new_cache[id] = table_clone(item)
    else
      log.warn("The IP address '", ip, "' failed to be added to the IP blacklist, ", err)
    end
  end
  raw = new_cache
  tree = new_tree
end


-- @tparam string ip
-- @return boolen, string [,string]
function _M.match(self, ip)
  return tree:match(ip)
end


function _M.reset(self)
  table_clear(raw)
  tree = ipmatcher.new()
end


function _M.get_raw(_)
  local cnt = {}
  for _, item in pairs(raw) do
    cnt[#cnt +1] = table_clone(item)
  end
  return cnt
end

return _M