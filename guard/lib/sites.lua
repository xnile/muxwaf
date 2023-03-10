local require        = require
local log            = require("log")
local utils          = require("utils")
local cjson          = require("cjson.safe")
local constants      = require("constants")
local ngx            = ngx
local ipairs         = ipairs
local table_new      = table.new
local table_clear    = table.clear
local table_deepcopy = table.deepcopy
local string_format  = string.format

local _M = {
    _VERSION = 0.1,
}

local ORIGIN_PROTOCOL = constants.ORIGIN_PROTOCOL

local sites        = table_new(0, 50)
local host_matcher = table_new(0, 50)
local cache = sites


-- TODO: add cache
local function get_site(host)
    local site_id = host_matcher[host]
    return site_id and sites[site_id] or nil
end


-- @schema
-- {
--     type = "array",
--     items = site_schema_def,
-- }

function _M.add(_, items)
    for _, item in ipairs(items) do
        local id, host = item.id, item.host

        if sites[id] then
            log.warn(string_format("faild add site: id '%s' already exists", id))
            goto continue
        end

        if host_matcher[host] then
            log.warn(string_format("failed add site: conflicting host '%s'", host))
            goto continue
        end

        sites[id] = table_deepcopy(item)
        host_matcher[host] = id
        log.info(string_format("add site '%s' success", host))
        ::continue::
    end
end

function _M.del(_, items)
    for _, id in ipairs(items) do
        local candidate = sites[id]
        if not candidate then
            log.warn(string_format("faild to delete site: id '%s' does not exist", id))
            goto continue
        end
        local host = candidate.host

        sites[id] =  nil
        host_matcher[host] = nil
        
        ::continue::
    end
end

function _M.update(_, items)
    for _, item in ipairs(items) do
        local id, host  = item.id, item.host   --TODO: remove host

        local candidate = cache[id]
        if not candidate then
            log.warn(string_format("faild to update site: id '%s' does not exist", id))
            goto continue
        end

        if host ~= candidate.host then
            log.warn("faild to update site: update host is not supported")
            goto continue
        end

        if item.config then
            candidate.config = table_deepcopy(item.config)
        end

        if item.origins then
            candidate.origins = table_deepcopy(item.origins)
        end

        ::continue::
    end
end


function _M.full_sync(_, items)
    local del_ids = utils.diff_cfg_ids(cache, items)
    local this = _M
    this:del(del_ids)

    for _, item in ipairs(items) do
        if not cache[item.id] then
            this:add({ item })
        else
            this:update({ item })
        end
    end
end

function _M.get_origin_protocol(host, request_scheme)
    local site = get_site(host)
    if not site then
        return "http"
    end

    local origin_protocol = site.config.origin_protocol

    if origin_protocol == ORIGIN_PROTOCOL.HTTP then
        return "http"
    elseif origin_protocol == ORIGIN_PROTOCOL.HTTPS then
        return "https"
    elseif origin_protocol == ORIGIN_PROTOCOL.FOLLOW then
        return request_scheme
    end
    return "http"
end

-- api for ctx
function _M.get_site_id(host)
    return host_matcher[host] and host_matcher[host] or ""
end

-- api for ssl
function _M.is_exist(host)
    return get_site(host) == nil and true or false
end

-- api for ssl
function _M.is_enable_https(host)
    local site = get_site(host)
    return site and site.config.is_https == 1 or false
end

-- api for ssl
function _M.get_site_cert_id(host)
    local site = get_site(host)
    return site and site.config.cert_id or ""
end

function _M.get_site_origins(host)
    local site = get_site(host)
    return site and site.origins or nil
end

function _M.is_real_ip_from_header(host)
    local site = get_site(host)
    return site and site.config.is_real_ip_from_header == 1 or false
end

function _M.get_real_ip_header(host)
    local site = get_site(host)
    return site and site.config.real_ip_header or "X-Forwarded-For" --TODO: move to constants
end

function _M.reset(_)
    table_clear(sites)
    table_clear(host_matcher)
end

function _M.get_raw(_)
  local cnt = {}
  for _, item in pairs(cache) do
    cnt[#cnt +1] = table_deepcopy(item)
  end
  return cnt
end

return _M
