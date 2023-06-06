-- TODO: support custom origin host
local require        = require
local balancer       = require("balancer")
local log            = require("log")
local utils          = require("utils")
local cjson          = require("cjson.safe")
local constants      = require("constants")
local ngx            = ngx
local ipairs         = ipairs
local table_new      = table.new
local table_clear    = table.clear
local table_clone    = require("table.clone")

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
            log.warn("failed to add site \"", host, "\" with ID \"", id, "\" the site already exists")
            goto continue
        end

        if host_matcher[host] then
            log.warn("failed to add site \"", host, "\" with ID \"", id, "\" conflicting host \"", host, "\"")
            goto continue
        end

        sites[id] = table_clone(item)
        host_matcher[host] = id
        log.debug("successed to add site \"", host, "\" with ID \"", id, "\"")
        ::continue::
    end
end

function _M.del(_, items)
    for _, id in ipairs(items) do
        local candidate = sites[id]
        if not candidate then
            log.warn("failed to delete site with ID \"", id, "\" the site does not exist")
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
            log.warn("failed to update site with ID \"", id, "\"the site does not exist")
            goto continue
        end

        if host ~= candidate.host then
            -- log.warn("failed to update site wh update host is not supported")
            log.warn("failed to update site with ID \"", id, "\" update site domain is not supported")
            goto continue
        end

        if item.config then
            candidate.config = table_clone(item.config)
        end

        if item.origins then
            candidate.origins = table_clone(item.origins)
        end

        ::continue::
    end
end


function _M.full_sync(_, items)

    balancer:full_sync_origins(items)

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
    return get_site(host) ~= nil and true or false
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
    cnt[#cnt +1] = table_clone(item)
  end
  return cnt
end

return _M
