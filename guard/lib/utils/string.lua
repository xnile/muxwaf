local string = string
local setmetatable = setmetatable


local _M = {
  _VERSION = 0.1
}

setmetatable(_M, {__index = string})


function _M.isempty(s)
  return s == nil or s == ''
end



return _M

