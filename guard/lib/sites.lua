local require        = require
local balancer       = require("balancer")
local log            = require("log")
local utils          = require("utils")
local cjson          = require("cjson.safe")
-- local constants      = require("constants")
local ngx            = ngx
local ipairs         = ipairs
local table_new      = table.new
local table_clear    = table.clear
local table_clone    = require("table.clone")

local _M = {
    _VERSION = 0.1,
}

-- local ORIGIN_PROTOCOL = constants.ORIGIN_PROTOCOL
local DEFAULT_REAL_IP_HEADER = "X-Forwarded-For"

local sites        = table_new(0, 20)
local host_matcher = table_new(0, 20)
local raw

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
    -- add orgins first
    -- balancer:add_origins(items)

    for _, item in ipairs(items) do
        local id, host, config = item.id, item.host, item.config

        if sites[id] then
            log.warn("failed to add site '", host, "' with ID '", id, "', the site already exists")
            goto continue
        end

        if host_matcher[host] then
            log.warn("failed to add site '", host, "' with ID '", id, "',  host conflict")
            goto continue
        end

        sites[id] = table_clone(item)
        host_matcher[host] = id

        balancer:add_origin_config(host, config.origin)

        log.debug("successed to add site '", host, "' with ID '", id, "'")

        ::continue::
    end
end

function _M.del(_, items)
    for _, id in ipairs(items) do
        local candidate = sites[id]
        if not candidate then
            log.warn("failed to delete site with ID '", id, "', the site does not exist")
            goto continue
        end
        local host = candidate.host

        sites[id] =  nil
        host_matcher[host] = nil

        balancer:del_origin_config(host)
        
        ::continue::
    end

    -- -- delete origins later
    -- balancer:del_origin_config(del_hosts)
end

function _M.update(_, items)
    -- balancer:update_origins(items)

    for _, item in ipairs(items) do
        local id, host, config  = item.id, item.host, item.config

        local candidate = sites[id]
        if not candidate then
            log.warn("failed to update site ", host, " with ID '", id, "', the site does not exist")
            goto continue
        end

        if host ~= candidate.host then
            -- log.warn("failed to update site wh update host is not supported")
            log.warn("failed to update site ", host, " with ID '", id, "', update site domain is not supported")
            goto continue
        end

        -- if item.config then
        --     candidate.config = table_clone(item.config)
        -- end

        -- if item.origins then
        --     candidate.origins = table_clone(item.origins)
        -- end

        for k,v in pairs(config) do
            candidate.config[k] = type(config[k]) == "table" and table_clone(config[k]) or config[k]
            if k == "origin" then
                balancer:update_origin_config(host, table_clone(config[k]))
            end
        end

        ::continue::
    end
end


function _M.full_sync(_, items)

    -- balancer:full_sync_origins(items)

    local del_ids = utils.diff_cfg_ids(sites, items)
    local this = _M
    this:del(del_ids)

    for _, item in ipairs(items) do
        if not sites[item.id] then
            this:add({ item })
        else
            this:update({ item })
        end
    end
end


-- TODO: move to balancer
-- function _M.get_origin_host(host)
--     local site = get_site(host)
--     return site and (site.config and (site.config.origin and site.config.origin.origin_host_header) or "") or ""
-- end

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
    return site and (site.config and site.config.is_https == 1 or false) or false
end

-- api for ssl
function _M.get_site_cert_id(host)
    local site = get_site(host)
    return site and (site.config and site.config.cert_id or false) or ""
end

function  _M.is_force_https(host)
    local site = get_site(host)
    return site and (site.config and site.config.is_force_https == 1 or false) or false
end

-- function _M.get_site_origins(host)
--     local site = get_site(host)
--     return site and site.origins or nil
-- end

function _M.is_real_ip_from_header(host)
    local site = get_site(host)
    return site and (site.config and site.config.is_real_ip_from_header == 1  or false) or false
end

function _M.get_real_ip_header(host)
    local site = get_site(host)
    return site and (site.config and site.config.real_ip_header or false) or DEFAULT_REAL_IP_HEADER
end

function _M.reset(_)
    table_clear(sites)
    table_clear(host_matcher)
    table_clear(raw)
end

function _M.get_raw(_)
    raw = table_clone(sites)

    local cnt = {}
    for _, item in pairs(raw) do
    cnt[#cnt +1] = table_clone(item)
    end
    return cnt
end

return _M