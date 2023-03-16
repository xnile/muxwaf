-- TODO: support for wildcard paths with other children
local require       = require
local cjson         = require("cjson.safe")
local table_new     = table.new
local table_insert  = table.insert
local table_sort    = table.sort
local str_len       = string.len
local str_sub       = string.sub
local ipairs        = ipairs
local error         = error
local string_format = string.format
local setmetatable  = setmetatable

local _M = {
    _VERSION = 0.1
}

local _mt = { __index = _M }

local PARAM = ":"

local function get_longest_common_prefix(a, b)
    local e = #a < #b and #a or #b

    local i, j = 1, 0
    while i <= e and a:sub(i, i) == b:sub(i, i) do
      i = i + 1
      j = j + 1
    end

    return j
end

local function find_param(path)
    local valid = true

    for i=1,#path,1 do
        if path:sub(i,i) ~= PARAM then
            goto continue
        end

        local e = path:sub(i+1)
        for j=1,#e,1 do
            local c = e:sub(j, j)
            if c == "/" then -- ahead
                return path:sub(i,i+j-1), i, valid
            elseif c == PARAM then
                valid = false
            end
        end
        do
            return path:sub(i), i, valid
        end

        ::continue::
    end
    
    return "", -1, false
end

local _ = {
    path = "/",
    value = 0,
    child = {},
    param = false,
    param_child = false,
    priority = 1,
    full_path = "/api/users",

}

local function insert_child(n, path, full_path, value)
    while true do
        local param, idx, valid = find_param(path)

        if idx == 1 then
        end

        if idx < 0 then
            break
        end

        if not valid then
            error(string_format("only one param per path segment is allowed, has: '%s in path %s'", full_path), 2)
        end

        if #param < 2 then
            error(string_format("param must be named with a non-empty name in path '%s'", full_path), 2)
        end

        if idx > 1 then
            n.path = path:sub(1, idx-1) -- TODO
            path = path:sub(idx)            
        end

        n.param_child = true
        local child = {
            path = param,
            value = value,
            priority = 0,
            param = true,
            full_path = full_path,
            indices = {},
            child = {},
        }

        n.child = n.child or {}
        -- n.child = {
        --     [PARAM] = child
        -- }
        n.child[PARAM] = child

        n = child
        n.priority = n.priority + 1


        if #param < #path then
            path = path:sub(#param+1)
            local child = {
                priority = 1,
                full_path = full_path,
                indices = {},
                child = {},
            }

            n.child = {
                ["/"] = child
            }

            n = child
            goto continue
        end

        do
            n.value = value
            return
        end

        ::continue::
    end

    n.path = path
    n.value = value
    n.full_path = full_path
end


local function incr_child_prio(n, c)
    local cnt = n.child[c]
    cnt.priority = n.priority +1

    table_sort(n.indices, function(a, b)
        if not n.child[a].priority then
            return false
        end

        if not n.child[b].priority then
            return true
        end

        return n.child[a].priority > n.child[b].priority
    end)

end


function _M.new()
    return setmetatable({}, _mt)
end


function _M.insert(self, path, value)
    if path:sub(1, 1) ~= "/" then
        return error("path should start with /")
    end

    local n = self

    local _full_path = path
    local _parent_full_path_idx = 0

    n.priority = (n.priority and n.priority or 0) + 1
    
    --  空节点
    if not n.path then
        -- n.path = path
        -- n.value = value
        -- n.full_path = path
        insert_child(n, path, full_path, value)
        return
    end

    while true do
        local idx = get_longest_common_prefix(path, n.path)

        -- split edge
        if idx < #n.path then
            local child = {
                path = str_sub(n.path, idx+1),
                value = n.value,
                child = n.child,
                priority = n.priority - 1,
                full_path = n.full_path,
                indices = n.indices or {},
                param_child = n.param_child,
            }

            local c = n.path:sub(idx+1, idx+1)
            n.child = n.child and n.child or table_new(0, 10) -- preallocated capacity
            n.child[c] = child
            n.indices = n.indices and n.indices or table_new(5, 0)
            table_insert(n.indices, c)
            n.path = path:sub(1, idx)
            n.full_path = str_sub(_full_path, 1, _parent_full_path_idx + idx)
            n.value = nil -- empty value
            n.param_child = false
            -- n.priority = n.priority   -- do nothing

        end

        -- eg: already existed /api/users then add /api/posts
        if idx < #path then  -- implies idx == #n.path
            path = path:sub(idx+1)

            if n.param_child then
                -- TODO
            end

            local c = path:sub(1, 1)
            if c ~= PARAM then        
                for _, i in ipairs(n.indices) do
                    if i == c then
                        incr_child_prio(n, i)
                        n = n.child[i] -- deep
                        goto continue
                    end
                end

                -- make new node
                table_insert(n.indices, c)

                local child = {
                    full_path = full_path,
                    indices = {},
                    child = {},
                }
                n.child[c] = child
                n = child
            end

            insert_child(n, path, full_path, value)

            return
        end
         

        do
            if n.value then
                return error(string_format("handler are already registered for path %s ", _full_path), 2)
            end

            n.value = value
            n.full_path = _full_path
            return
        end

        ::continue::
    end
end

function _M.get(self, path)
    local n = self
    local param = {}
    while true do
        local prefix = n.path

        -- bug fix for empty tree
        if not prefix then
            return nil
        end

        if #path > #prefix then
            if path:sub(1, #prefix) == prefix then -- prefix same
                path = path:sub(#prefix +1)

                if n.param_child then

                    -- exact math
                    local idxc = path:sub(1, 1)
                    for _, c in ipairs(n.indices) do
                        if idxc == c then
                            n = n.child[c]
                            goto continue
                        end
                    end

                    n = n.child[PARAM]

                    local last = 0
                    for i=1,#path,1 do
                        if path:sub(1,1) ~= "/" then
                            last = last + 1
                        end
                    end

                    param[n.path:sub(2)] = path:sub(1, last) -- param

                    if last < #path then
                        if n.child["/"] then
                            n = n.child["/"]
                            goto continue
                        end
                        return  -- not found
                    end

                    if n.value then
                        return n.value, param
                    end
                end


                local idxc = path:sub(1, 1)
                for _, c in ipairs(n.indices) do
                    if idxc == c then
                        n = n.child[c]
                        goto continue
                    end
                end

                return nil
            end
        end

        if path == prefix then
            if n.value then
                return n.value, n.full_path
            end
        end

        do
            return nil
        end
    ::continue::
    end
end

return _M
