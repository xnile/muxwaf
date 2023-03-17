local string   = string
local str_byte = string.byte
local setmetatable = setmetatable


local _M = {
  _VERSION = 0.1
}

setmetatable(_M, {__index = string})


function _M.isempty(s)
  return s == nil or s == ''
end

function _M.rfind_char(s, ch, idx)
    local b = str_byte(ch)
    for i = idx or #s, 1, -1 do
        if str_byte(s, i, i) == b then
            return i
        end
    end
    return nil
end


return _M

