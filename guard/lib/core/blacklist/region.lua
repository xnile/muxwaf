local log           = require("log")
local table_new     = table.new
local table_clear   = table.clear

local _M  = {
  _VERSION = 0.1,
}

local matcher = table_new(0, 50)
local empty_table = {}
local MATCH_MODE = {
    BLACKLIST = 0,
    WHITELIST = 1,
}

local function update_with_add(items)
    if not items then return end
    for _, item in ipairs(items) do
        local site_id = item.site_id
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

        log.debug("successed to update region IP blacklist of site \"", site_id, "\"")
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
            log.warn("failed to clear region IP blacklist of site \"", site_id, "\" site does not exist")
            goto continue
        end

        log.debug("successed to clear region IP blacklist of site \"", site_id, "\"")
        matcher[site_id] = nil

        ::continue::
    end
end


function _M.full_sync(_, items)
    table_clear(matcher)  --TODO: diff and delete
    update_with_add(items)
end

-- @param site_id string
-- @param ip string
-- @return boolean
function _M.match(ctx)
    local site_id = ctx.site_id
    local candidate = matcher[site_id]
    if not candidate then
        return false
    end

    local location = ctx.ip_location
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