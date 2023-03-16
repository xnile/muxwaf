local table = table
local table_new = table.new
local table_insert = table.insert
local ipairs = ipairs
local pairs  = pairs


local _M = {
    version = 0.1,
}

setmetatable(_M, {__index = table})

function _M.array_merge(tabA, tabB)
  local lenA, lenB = #tabA, #tabB
  local tab = table.new(lenA + lenB, 0)

  for i = 1, lenA do
    tab[i] = tabA[i]
  end
  for i = 1, lenB do
    tab[lenA+i] = tabB[i]
  end

  return tab
end


function _M.array_contains(t, e)
  local r = false

  for _, v in ipairs(t) do
    if e == v then
      r = true
    end
  end

  return r
end


-- function _M.is_empty(t)
--     return t == nil or next(t) == nil
-- end


-- from http://lua-users.org/wiki/CopyTable
-- function _M.deepcopy(orig)
--     local orig_type = type(orig)
--     local copy
--     if orig_type == 'table' then
--         copy = {}
--         for orig_key, orig_value in next, orig, nil do
--             copy[_M.deepcopy(orig_key)] = _M.deepcopy(orig_value)
--         end
--         setmetatable(copy, _M.deepcopy(getmetatable(orig)))
--     else -- number, string, boolean, etc
--         copy = orig
--     end
--     return copy
-- end


return _M
