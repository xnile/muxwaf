local require        = require
local log            = require("log")
local utils          = require("utils")
local radix          = require("resty.radixtree")
local ratelimit      = require("resty.ratelimit")
local pairs          = pairs
local ipairs         = ipairs
local setmetatable   = setmetatable
local table_new      = table.new
local table_clear    = table.clear
local table_clone    = require("table.clone")

local dict_name         = require("constants").DICTS.RATELIMIT
local ratelimit_device  = ratelimit.new(dict_name)

local cache  = table_new(0, 100)
local tree = table_new(0, 20)

local _M = {}

--  mode 1: prefix match
--       2: exact match
function _M.add(self, items)

  for _, item in pairs(items) do
    local id, host, path, match_mode = item.id, item.host, item.path, item.match_mode

    cache[id] = table_clone(item)

    if not tree[host] then
      tree[host] = {
        prefix = radix.new(),
        exact  = radix.new(),
      }
    end

    -- TODO: defined by constant
    if match_mode == 1 then -- prefix
      local ok, err = tree[host].prefix:insert(path, id)
      if not ok then
        log.error("failed to add rate limit with ID \"", id, "\", ", err)
      end
      log.debug("successed to add rate limit, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    elseif match_mode == 2 then -- exact
      local ok,err = tree[host].exact:insert(path, id)
      if not ok then
        log.error("failed to add rate limit with ID \"", id, "\", ", err)
      end
      log.debug("successed to add rate limit, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    else
      log.warn("failed to add rate limit, unsupported match mode \"", match_mode, "\"")
      goto continue
    end
  end

  ::continue::
end


function _M.del(self, items)

  for _, id in pairs(items) do
    local rule = cache[id]
    if not rule then
      log.warn("faild to delete rate limit, the rule with ID \'", id, "\' does not exist")
      goto continue
    end

    local host, path, match_mode = rule.host, rule.path, rule.match_mode
    if match_mode == 1 then
      local ok, err = tree[host].prefix:remove(path)
      if not ok then
        log.error("failed to delete rate limit with ID \"", id, "\", ", err)
      end
      log.debug("successed to delete rate limit, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    elseif match_mode == 2 then
      local ok, err = tree[host].exact:remove(path)
      if not ok then
        log.error("failed to delete rate limit with ID \"", id, "\", ", err)
      end
      log.debug("successed to delete rate limit, id \"", id, "\", host \"", host, "\", URL \"", path, "\", match mode \"", match_mode, "\"")
    else
      log.debug("failed to delete rate limit: match mode \'", match_mode, "\' not supported")
      goto continue
    end

    cache[id] = nil
    log.debug("succeeded to delete the rate limit rule with ID \'", id, "\'")
    ::continue::
  end
end


function _M.update(self, items)
  for _, item in ipairs(items) do

    local id, host, path, match_mode = item.id, item.host, item.path, item.match_mode
    local old = cache[id]
    if not old then
      log.warn("faild to update rate limit, the rule with ID \'", id, "\' does not exist")
      goto continue
    end

    -- only limit or window changed
    if host == old.host and path == old.path and match_mode == old.match_mode then
      cache[id] = table_clone(item)
      goto continue
    end

    -- TODO: Do not reuse the logic
    local this = _M
    this:del({ id })
    this:add({ item })

    ::continue::
  end
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


function _M.url_match(self, host, path)
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

function _M.incomming(self, id, key)
  -- the rule has already been deleted
  if not cache[id] then
    log.debug("The rule with id '", id, "'  has already been deleted")
    return ok, nil
  end

  local limit, window = cache[id].limit, cache[id].window

  return ratelimit_device:incomming(key, limit, window)
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
  ratelimit_device:flush_all()
end

return _M