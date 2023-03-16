local cjson = require("cjson.safe")
local setmetatable = setmetatable
local table = require("utils.tablex")
local ipairs = ipairs
local pairs = pairs
local error = error


local _M = {
    _VERSION = 0.1
}

setmetatable(_M, {__index = cjson})

function _M.validator(data, schema, hopeArry)
    if type(schema) ~= "table" then
        error("schema should be a table", 2)
        return
    end

    local ok, err = true, {}
    if hopeArry then
        if not table.is_array(data) then
            ok, err[#err+1] = false, "data should be a array"
            return ok, err
        end

        for i,t in ipairs(data) do
            if #schema < #t then
                for k,v in pairs(t) do
                    if not schema[k] then
                        ok, err[#err+1]= false, "the field "..k.." of element "..i.." cannot be present"
                    end
                    if type(v) ~= schema[k] then
                        ok, err[#err+1]= false, "the field "..k.." of element "..i.." should be a "..schema[k]
                    end
                end
            else
                for k,v in pairs(schema) do
                    if not t[k] then
                        ok, err[#err+1]= false, "the field "..k.." of element "..i.." should be present"
                    end
                    if type(t[k]) ~= schema[k] then
                        ok, err[#err+1]= false, "the field "..k.." of element "..i.." should be a "..schema[k]
                    end
                end
            end
        end
        return ok, err
    else
        if table.is_array(data) then
            ok, err[#err+1] = false, "data cannot be a array"
            return ok, err
        end
    end

    if #schema < #data then
        for k,v in pairs(data) do
            if not schema[k] then
                ok, err[#err+1]= false, "field "..k.." cannot be present"
            end
            
            if type(v) ~= schema[k] then
                ok, err[#err+1]= false, "field "..k.." should be a "..schema[k]
            end
        end
    else
        for k,v in pairs(schema) do
            if not data[k] then
                ok, err[#err+1] = false, "field "..k.." should be present"
            end
            if type(data[k]) ~= schema[k] then
                ok, err[#err+1] = false, "field "..k.." should be a "..schema[k]
            end
        end 
    end

    return ok, err
end


return _M