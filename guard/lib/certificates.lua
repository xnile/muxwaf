local require        = require
local log            = require("log")
local utils          = require("utils")
local constants      = require("constants")
local cjson          = require("cjson.safe")
local ngx            = ngx
local ngx_shared     = ngx.shared
local ipairs         = ipairs
local string_format  = string.format
local table_new      = table.new
local table_clear    = table.clear
local table_clone    = require("table.clone")

-- local shm       = ngx_shared[constants.DICTS.CERTIFICATE]
local certificates = table_new(0, 20)
local cache        = certificates

local _M = {
    
}

function _M.add(_, items)
    for _, item in ipairs(items) do
        local id, cert, key = item.id, item.cert, item.key
        if certificates[id] then
            log.warn(string_format("faild add certificate: id '%s' already exist", id))
            goto continue
        end
        certificates[id] = {
            cert = cert,
            key  = key,
        }

        ::continue::
    end
end


function _M.del(_, items)
    for _, id in ipairs(items) do
        if not certificates[id] then
            log.warn(string_format("faild to delete certificate: id '%s' dose not exist", id))
            goto continue
        end
        certificates[id] = nil

        ::continue::
    end
end


function _M.update(_, items)
    for _, item in ipairs(items) do
        local id, cert, key = item.id, item.cert, item.key
        if not certificates[id] then
            log.warn(string_format("faild to update certificate: id '%s' dose not exist", id))
            goto continue
        end

        certificates[id] = {
            cert = cert,
            key  = key,
        }

        ::continue::
    end
end

function _M.full_sync(_, items)
    -- local del_ids = utils.diff_cfg_ids(table_clone(cache), items)

    -- local this = _M
    -- this:del(del_ids)

    -- for _, item in ipairs(items) do
    --     if not cache[item.id] then
    --         this:add({ item })
    --     else
    --         this:update({item })
    --     end
    -- end
    local new_certificates = table_new(0, 20)
    for _, item in ipairs(items) do
        local id, cert, key = item.id, item.cert, item.key
        new_certificates[id] = {
            cert = cert,
            key  = key,
        }
    end
    certificates = new_certificates
    cache = certificates
    log.info("full sync certificates success")
end

function _M.get(id)
    local cert = certificates[id]
    if not cert then
        return nil, string_format("the certificate id '%s' dose not exist", id)
    end
    -- TODO: add cache
    return cert.cert .. "\n" .. cert.key, nil
end

function _M.reset(_)
    table_clear(certificates)
end

function _M.get_raw(_)
    local cnt = {}
    for id, item in pairs(cache) do
        cnt[#cnt+1] = {
            id = id,
            cert = item.cert,
            key  = item.key,
        }
    end
    return cnt
end

return _M