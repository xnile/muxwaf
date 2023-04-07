local table_new = table.new

local tmp_table = {}

local _M = {
  _VERSION = 0.1
}

function _M.pretty_bytes(c)
    if c > 1024 * 1024 then
        return string.format('%.2f', c / 1024 / 1024) .. 'M'
    elseif c > 1024 then
        return string.format('%.2f', c / 1024) .. 'K'
    else
        return string.format('%.2f', c) .. 'B'
    end
end

function _M.pretty_bandwidth(c)
    if c > 1024 * 1024 then
        return string.format('%.2f', c / 1024 / 1024) .. 'Mbps'
    elseif c > 1024 then
        return string.format('%.2f', c / 1024) .. 'Kbps'
    else
        return string.format('%.2f', c) .. 'bps'
    end
end


function _M.pretty_number(num)
    if not num then return 0 end
    if math.abs(num) < 1000 then return num end
    local neg = num < 0 and "-" or ""
    local left, mid, right = tostring(math.abs(num)):match("^([^%d]*%d)(%d*)(.-)$")
    return ("%s%s%s%s"):format(neg, left, mid:reverse():gsub("(%d%d%d)", "%1,"):reverse(), right)
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