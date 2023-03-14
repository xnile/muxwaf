local table_new = table.new

local tmp_table = {}

local _M = {
  _VERSION = 0.1
}

function _M.format_capacity(c)
    if c > 1024 * 1024 then
        return string.format('%.2f', c / 1024 / 1024) .. 'M'
    elseif c > 1024 then
        return string.format('%.2f', c / 1024) .. 'K'
    else
        return string.format('%.2f', c) .. 'B'
    end
end


-- @param old table, kv like table
-- @param new table, array like table
function _M.diff_cfg_ids(old, new)
  local ids = table_new(0, #new)
  for _, v in ipairs(new) do
    if v.id then
      ids[v.id] = tmp_table
    end
  end

  local diff_ids = {}
  for id, _ in pairs(old) do
    if not ids[id] then
      diff_ids[#diff_ids+1] = id
    end
  end

  return diff_ids
end

return _M