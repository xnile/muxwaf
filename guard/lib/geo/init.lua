local lrucache    = require("resty.lrucache.pureffi")
local ipip_parser = require("geo.parser.ipip.city")
local xdb_parser  = require("geo.parser.ip2region.xdbffi")
local net = require("utils.net")


local _M = {
    _VERSION = 0.1
}

local IPIPDB_PATH  = ngx.config.prefix() .. "ipdb/ipipfree.ipdb"
local XDB_PATH     = ngx.config.prefix() .. "ipdb/ip2region.xdb"

local IP_GEO_CACHE_SIZE = 1000
local IP_GEO_CACHE_TTL  = 3600

local EMPTY_IP_GEO =  {
    country_name = "",
    city_name = "",
    region_name = "",
}

local providers = {
    ipip = {
        parser = ipip_parser,
        db_path = IPIPDB_PATH,
        db_fields = {"city_name", "country_name", "region_name"},
    },
    xdb = {
        parser = xdb_parser,
        db_path = XDB_PATH,
        db_fields = {
            "continent",
            "country_name",
            "region_name",
            "city_name",
            "district_name",
            "china_admin_code1",
            "china_admin_code2",
            "china_admin_code3",
            "country",
            "country_code",
            "idd_code",
            "isp",
            "lat",
            "lng",
        }
    }
}

local _mt = { __index = _M }

local ip_geo_cache, err = lrucache.new(IP_GEO_CACHE_SIZE)
if not ip_geo_cache then
    error("failed to create the cache: " .. (err or "unknown"), 2)
end


function _M.new(provider)
    local candidate = providers[provider]
    if not candidate then
        error("'" .. provider .. "' is not supported ip database", 2)
    end

    local db
    if provider == "ipip" then
        db = candidate.parser.new(nil, candidate.db_path)
    else
        db = candidate.parser.new(candidate.db_path)
    end

    local self = {
        -- db = candidate.new(db_path[provider]),
        db = db,
        provider = provider,
    }

    return setmetatable(self, _mt)
end

local function parse_region_city(name)
    local name, _, _ = ngx.re.gsub(name, "省|市", "", "jo")
    return name
end


local function validate_ip(ip)
    if type(ip) ~= "string" then
        return false, "ip should a string"
    end

    if net.is_valid_ipv6(ip) then
        return nil, "IPv6 is not currently supported"
    end

    if not net.is_valid_ip(ip) then
        return nil, "'" .. ip .. "' is an invalid ipv4"
    end

    return true, nil
end

function _M.search(self, ip)
    local ok ,err = validate_ip(ip)
    if not ok then
        return EMPTY_IP_GEO, err
    end

    local geo = ip_geo_cache:get(ip)
    if not geo then
        local db, provider = self.db, self.provider

        if provider == "ipip" then
            local res = db:find(ip, "CN")
            geo = {
                country_name = res.country_name,
                region_name = res.region_name,
                city_name = res.city_name
            }


        elseif provider == "xdb" then
            local res, err = db:search(ip)
            if not res then
                return EMPTY_IP_GEO, err
            end

            geo = {
                country_name = res.country_name,
                region_name = parse_region_city(res.region_name),
                city_name = parse_region_city(res.city_name)
            }
        end
        ip_geo_cache:set(ip, geo, IP_GEO_CACHE_TTL)
    end

    return geo, nil
end


return _M