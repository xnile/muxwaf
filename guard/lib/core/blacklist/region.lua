local lrucache      = require "resty.lrucache.pureffi"
local log           = require("log")
local table_new     = table.new
local table_clear   = table.clear
local string_format = string.format
-- local table_clone    = require("table.clone")

local _M  = {
  _VERSION = 0.1,
}

local matcher = table_new(0, 50)
local empty_table = {}
local MATCH_MODE = {
    BLACKLIST = 0,
    WHITELIST = 1,
}

local ip_geo_cache, err = lrucache.new(1000)
if not ip_geo_cache then
    error("failed to create the cache: " .. (err or "unknown"), 2)
end

local ipdb
local function find_ip_location(ip)
    if not ipdb then
        ipdb = muxwaf.get_ipdb()
    end
    local location = ip_geo_cache:get(ip)
    if not location then
        location = ipdb.ipip:find(ip, "CN")
        ip_geo_cache:set(ip, location, 3600)
    end
    return location
end

local function update_with_add(items)
    if not items then return end
    for _, item in ipairs(items) do
        if not matcher[item.site_id] then
            matcher[item.site_id] = table_new(0, 5)
        end

        local candidate = matcher[item.site_id]
        candidate.match_mode = item.match_mode
        candidate.site_id = item.site_id

        local _countries = item.countries and item.countries or {}
        local new_countries = {}
        for _, country in pairs(_countries) do
            new_countries[country] = empty_table
        end
        candidate.countries = new_countries

        local _regions = item.regions and item.regions or {}
        local new_regions = {}
        for _, region in pairs(_regions) do
            new_regions[region] = empty_table
        end
        candidate.regions = new_regions
    end
end

function _M.add(_, items)
    update_with_add(items)
end

function _M.update(_, items)
    update_with_add(items)
end


function _M.del(_, items)
    for _, site_id in ipairs(items) do
        if not matcher[site_id] then
            log.warn(string_format("faild to delete region blacklist: '%s' does not exist", site_id))
            goto continue
        end

        matcher[site_id] = nil
        log.info(string_format("delete region blacklist '%s' success", site_id))
        ::continue::
    end
end


function _M.full_sync(_, items)
    table_clear(matcher)  --TODO: diff and delete
    update_with_add(items)
    log.info("full sync region blacklist success")
end

-- @param site_id string
-- @param ip string
-- @return boolean
function _M.match(site_id, ip)
    local candidate = matcher[site_id]
    if not ip or not candidate then
        return false
    end

    -- local ipdb = muxwaf.get_ipdb()
    -- local location = ipdb.ipip:find(ip, "CN")
    local location = find_ip_location(ip)
    local country = location.country_name
    local region  = location.region_name

    if candidate.countries[country] then
        if candidate.mode == MATCH_MODE.WHITELIST then
            return false
        end
        return true
    end
    if candidate.regions[region] then
        if candidate.mode == MATCH_MODE.WHITELIST then
            return false
        end
        return true
    end
    return false
end

function _M.reset()
    table_clear(matcher)
end

function _M.get_raw()
    local raw = {}
    for _, v in pairs(matcher) do
        local item = table_new(0, 4)
        item.site_id = v.site_id
        item.match_mode = v.match_mode
        item.countries = {}
        item.regions = {}

        for country, _ in pairs(v.countries) do
            item.countries[#item.countries +1] = country
        end

        for region, _ in pairs(v.regions) do
            item.regions[#item.regions +1] = region
        end

        raw[#raw +1] = item
    end
    return raw
end

return _M