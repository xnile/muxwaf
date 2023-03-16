local table_new    = table.new
local table_insert = table.insert
local table_remove = table.remove
local setmetatable = setmetatable

local _M = {
    _VERSION = 0.1
}

function _M.new(size)
    local size = size or 0
    local self = table_new(size, 0)
    local reverse = table_new(0, size)

    return setmetatable(self, {
        __index = {
            insert = function(self, elem)
                -- print(elem)
                if not reverse[elem] then
                    table_insert(self, elem)
                    reverse[elem] = #self
                end
            end,

            remove = function(self, elem)
                local idx = reverse[elem]
                if idx then
                    reverse[elem] = nil
                    local end_elem = table_remove(self)
                    if end_elem ~= elem then
                        reverse[end_elem] = idx
                        self[idx] = end_elem
                    end
                end
            end,

            contains = function(self, elem)
                return reverse[elem] ~= nil
            end
        }
    })

end


return _M