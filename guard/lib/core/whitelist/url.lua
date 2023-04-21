local require        = require
local radix          = require("resty.radixtree")
local utils          = require("utils")
local log            = require("log")
local setmetatable   = setmetatable
local tostring       = tostring
local tonumber       = tonumber
local ipairs         = ipairs
local table_new      = table.new
local table_clear    = table.clear
local table_clone    = require("table.clone")

local cache  = table_new(0, 50)
local tree = table_new(0, 10)

local _M = {}

-- @tparam table items
function _M.add(self, items)
  for _, item in ipairs(items) do
    local id, host, path, match_mode = item.id, item.host, item.path, item.match_mode

    if cache[id] then
      log.warn("failed add URL whitelist, the rule id \"", id, "\" already exists")
      goto continue
    end    

    if not tree[host] then
      tree[host] = {
        prefix = radix.new(),
        exact  = radix.new(),
      }
    end

    if match_mode == 1 then -- prefix
      tree[host].prefix:insert(path, id)
    elseif match_mode == 2 then -- exact
      tree[host].exact:insert(path, id)
    else
      log.warn("unsupported match mode \"", match_mode, "\"")
      goto continue
    end

    cache[id] = table_clone(item)
    log.debug("successed to add url whitelist, host is \"", host, "\", URL is \"", path, "\", matching mode is \"", match_mode, "\"")
  end

  ::continue::
end


function _M.del(self, items)
  for _, id in ipairs(items) do
    local rule = cache[id]
    if not rule then
      log.warn("failed to delete URL whitelist, the rule with ID \"", id, "\" does not exist")
      goto continue
    end

    local host, path, match_mode = rule.host, rule.path, rule.match_mode
    if match_mode == 1 then
      local ok, err = tree[host].prefix:remove(path)
      if not ok then
        log.error("Failed to delete the URL whitelist with ID \"", id, "\"", err)
      end
      log.debug("successed to delete the URL whitelist, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    elseif match_mode == 2 then
      local ok, err = tree[host].exact:remove(path)
      if not ok then
        log.error("Failed to delete the URL whitelist with ID \"", id, "\"", err)
      end
      log.debug("successed to delete the URL whitelist, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    else
      log.error("Failed to delete the URL whitelist with ID \"", id, "\", the match mode \"", tostring(match_mode), " does not supported")
      goto continue
    end

    cache[id] = nil
    ::continue::
  end
end


function _M.update(self, items)
  for _,item in ipairs(items)do
    local id = item.id
    if not cache[id] then
      log.warn("failed to update URL whitelist, the rule with ID \"", id, "\" does not exist")
      goto continue
    end

    -- TODO: Do not reuse the logic
    local this = _M
    this:del({ id })
    this:add({ item })

    ::continue::
  end
  return
end


function _M.full_sync(_, items)
  local del_ids = utils.diff_cfg_ids(table_clone(cache), items)

  local this = _M
  this:del(del_ids)

  for _, item in ipairs(items) do
    if not cache[item.id] then
      this:add({ item })
    else
      this:update({item })
    end
  end
end


function _M.match(self, host, path)
  if not tree[host] then
    return nil
  end

  -- first exact search
  local id, _ = tree[host].exact:exact_find(path)
  if id then
    return id
  end

  local id, _ = tree[host].prefix:prefix_find(path) 
  if id then 
    return id
  end  

  return nil
end


function _M.get_raw(_)
  local cnt = {}
  for _, item in pairs(cache) do
    cnt[#cnt +1] = table_clone(item)
  end
  return cnt
end


function _M.reset(self)
  table_clear(cache)
  table_clear(tree)
end


return _M
