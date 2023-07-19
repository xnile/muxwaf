local log         = require("log")
local bit         = require("bit")
local clear_tab   = require("table.clear")
local nkeys       = require("table.nkeys")
local new_tab     = table.new
local str_find    = ngx.re.find
local tonumber    = tonumber
local ipairs      = ipairs
local pairs       = pairs
local ffi         = require "ffi"
local ffi_cdef    = ffi.cdef
local ffi_copy    = ffi.copy
local ffi_new     = ffi.new
local C           = ffi.C
local insert_tab  = table.insert
local sort_tab    = table.sort
local string      = string
local setmetatable=setmetatable
local type        = type
local error       = error
local str_sub     = string.sub
local str_byte    = string.byte

local AF_INET     = 2
local AF_INET6    = 10
if ffi.os == "OSX" then
    AF_INET6 = 30
end

local _M = {_VERSION = 0.3}
local mt = {__index = _M}


ffi_cdef[[
    int inet_pton(int af, const char * restrict src, void * restrict dst);
    uint32_t ntohl(uint32_t netlong);
]]

local get_ip_family
do
    local inet = ffi_new("unsigned int [1]")

    function get_ip_family(ip)
        if C.inet_pton(AF_INET, ip, inet) == 1 then
            return AF_INET
        end

        if C.inet_pton(AF_INET6, ip, inet) == 1 then
            return AF_INET6
        end

        return false
    end
end

local is_valid_ip_or_cidr
do
    function is_valid_ip_or_cidr(ip)
        local mask = 0
        local pos = str_find(ip, "/", "jo")
        if pos then
            mask = tonumber(str_sub(ip, pos + 1))
            ip = str_sub(ip, 1, pos - 1)
        end

        if mask < 0 or mask > 128 then
            return false
        end

        local family = get_ip_family(ip)
        if not family then
            return false
        end

        if family == AF_INET and mask > 32 then
            return false
        end

        return true
    end
end


local parse_ipv4
do
    local inet = ffi_new("unsigned int [1]")

    function parse_ipv4(ip)
        if not ip then
            return false
        end

        if C.inet_pton(AF_INET, ip, inet) ~= 1 then
            return false
        end

        return C.ntohl(inet[0])
    end
end
_M.parse_ipv4 = parse_ipv4

local parse_ipv6
do
    local inets = ffi_new("unsigned int [4]")

    function parse_ipv6(ip)
        if not ip then
            return false
        end

        if str_byte(ip, 1, 1) == str_byte('[')
            and str_byte(ip, #ip) == str_byte(']') then

            -- strip square brackets around IPv6 literal if present
            ip = str_sub(ip, 2, #ip - 1)
        end

        if C.inet_pton(AF_INET6, ip, inets) ~= 1 then
            return false
        end

        local inets_arr = new_tab(4, 0)
        for i = 0, 3 do
            insert_tab(inets_arr, C.ntohl(inets[i]))
        end
        return inets_arr
    end
end
_M.parse_ipv6 = parse_ipv6


local function split_ip(ip_addr_org)
    -- local idx = str_find(ip_addr_org, "/", 1, true)
    local idx = str_find(ip_addr_org, "/", "jo")
    if not idx then
        return ip_addr_org
    end

    local ip_addr = str_sub(ip_addr_org, 1, idx - 1)
    local ip_addr_mask = str_sub(ip_addr_org, idx + 1)
    return ip_addr, tonumber(ip_addr_mask)
end
-- _M.split_ip = split_ip


local idxs = {}
local function gen_ipv6_idxs(inets_ipv6, mask)
    clear_tab(idxs)

    for _, inet in ipairs(inets_ipv6) do
        local valid_mask = mask
        if valid_mask > 32 then
            valid_mask = 32
        end

        if valid_mask == 32 then
            insert_tab(idxs, inet)
        else
            insert_tab(idxs, bit.rshift(inet, 32 - valid_mask))
        end

        mask = mask - 32
        if mask <= 0 then
            break
        end
    end

    return idxs
end


local function cmp(x, y)
    return x > y
end



local function new()
    return setmetatable({
        ipv4 = {},
        ipv4_mask = {},
        ipv4_mask_arr = {},
        ipv4_match_all_value = nil,  -- TODO: remove it

        ipv6 = {},
        ipv6_mask = {},
        ipv6_mask_arr = {},
        ipv6_values = {},
        ipv6s_values_idx = 1,
        ipv6_match_all_value = nil, -- TODO: remove it
    }, mt)
  end

_M.new = new

-- TODO: deep delete
local insert_ipv4
do
  function insert_ipv4(self, inet_ipv4, ip_addr_mask, value)
    local parsed_ipv4s = self.ipv4
    local parsed_ipv4s_mask = self.ipv4_mask 
    local ipv4_match_all_value = self.ipv4_match_all_value

    ip_addr_mask = ip_addr_mask or 32
    if ip_addr_mask == 32 then
      parsed_ipv4s[inet_ipv4] = value
    elseif ip_addr_mask == 0 then
      ipv4_match_all_value = value
    else
      local valid_inet_addr = bit.rshift(inet_ipv4, 32 - ip_addr_mask)
      -- print(valid_inet_addr)           
      parsed_ipv4s_mask[ip_addr_mask] = parsed_ipv4s_mask[ip_addr_mask] or {}
      parsed_ipv4s_mask[ip_addr_mask][valid_inet_addr] = value
      log.debug("ipv4 mask: ", ip_addr_mask,
                " valid inet: ", valid_inet_addr)
    end

    local ipv4_mask_arr = new_tab(nkeys(parsed_ipv4s_mask), 0)
    local i = 1
    for k, _ in pairs(parsed_ipv4s_mask) do
        ipv4_mask_arr[i] = k
        i = i + 1
    end
    sort_tab(ipv4_mask_arr, cmp)

    self.ipv4_mask_arr = ipv4_mask_arr

    return true
  end
end


local insert_ipv6
do
    function insert_ipv6(self, ip_addr, inets_ipv6, ip_addr_mask, value)
        local parsed_ipv6s = self.ipv6
        local parsed_ipv6s_mask = self.ipv6_mask
        local ipv6_values = self.ipv6_values
        local ipv6s_values_idx = self.ipv6s_values_idx
        local ipv6_match_all_value = self.ipv6_match_all_value

        ip_addr_mask = ip_addr_mask or 128
        if ip_addr_mask == 128 then
            parsed_ipv6s[ip_addr] = value

        elseif ip_addr_mask == 0 then
            ipv6_match_all_value = value
        end

        parsed_ipv6s[ip_addr_mask] = parsed_ipv6s[ip_addr_mask] or {}

        local inets_idxs = gen_ipv6_idxs(inets_ipv6, ip_addr_mask)
        local node = parsed_ipv6s[ip_addr_mask]
        for i, inet in ipairs(inets_idxs) do
            if i == #inets_idxs then
                if value then
                    ipv6_values[ipv6s_values_idx] = value
                    node[inet] = ipv6s_values_idx
                    ipv6s_values_idx = ipv6s_values_idx + 1
                -- the case for the with_value is nil
                -- else
                --     node[inet] = true
                end
            end
            node[inet] = node[inet] or {}
            node = node[inet]
        end

        parsed_ipv6s_mask[ip_addr_mask] = true


        local ipv6_mask_arr = new_tab(nkeys(parsed_ipv6s_mask), 0)
        local i = 1
        for k, _ in pairs(parsed_ipv6s_mask) do
            ipv6_mask_arr[i] = k
            i = i + 1
        end
        sort_tab(ipv6_mask_arr, cmp)

        self.ipv6_mask_arr = ipv6_mask_arr

        return true
    end
end


-- TODO: deep remove
local remove_ipv4
do
    local err_msg = "ip or cidr does not exist"
    function remove_ipv4(self, inet_ipv4, ip_addr_mask, force)
        local ipv4_mask = self.ipv4_mask
        local parsed_ipv4s = self.ipv4
        local ipv4_match_all_value = self.ipv4_match_all_value

        ip_addr_mask = ip_addr_mask or 32

        if ip_addr_mask == 32 then
            local value = parsed_ipv4s[inet_ipv4]
            if not force and not value then
                return false, err_msg
            end
            parsed_ipv4s[inet_ipv4] = nil
            return true
        end

        if ip_addr_mask == 0 then
            if not force and not ipv4_match_all_value then
                return false, err_msg
            end
            ipv4_match_all_value = nil
            return true
        end

        if not ipv4_mask[ip_addr_mask] then
            if not force then
                return false, err_msg
            end
            return true
        end

        local valid_inet_addr = bit.rshift(inet_ipv4, 32 - ip_addr_mask)
        
        local value = ipv4_mask[ip_addr_mask][valid_inet_addr]
        if not force and not value then
            return false, err_msg
        end

        ipv4_mask[ip_addr_mask][valid_inet_addr] = nil
        return true
    end
end


-- TODO: deep remove
local remove_ipv6
do
    local err_msg = "ip or cidr does not exist"
    function remove_ipv6(self, ip_addr, inets_ipv6, ip_addr_mask, force)
        local parsed_ipv6s = self.ipv6
        local ipv6_values = self.ipv6_values
        local ipv6_match_all_value = self.ipv6_match_all_value
        
        ip_addr_mask = ip_addr_mask or 128

        if ip_addr_mask == 128 then
            local value = parsed_ipv6s[ip_addr]
            if not force and not value then
                return false, err_msg
            end
            parsed_ipv6s[ip_addr] = nil
            return true
        elseif ip_addr_mask == 0 then
            if not force and not ipv6_match_all_value then
                return false, err_msg
            end
            ipv6_match_all_value = nil
            return true
        end

        if not parsed_ipv6s[ip_addr_mask] then
            if not force then
                return false, err_msg
            end
            return true
        end


        local inets_idxs = gen_ipv6_idxs(inets_ipv6, ip_addr_mask)
        local node = parsed_ipv6s[ip_addr_mask]
        for i, inet in ipairs(inets_idxs) do
            if i == #inets_idxs then
                local ipv6s_values_idx = node[inet]
                if not ipv6s_values_idx then
                    if not force then
                        break
                    end
                    return true
                end
                ipv6_values[ipv6s_values_idx] = nil
                node[inet] = nil
                return true
            end

            if not node[inet] then
                if not force then
                    break
                end
                return true
            end

            node = node[inet]
        end

        return false, err_msg
    end
end

-- @tparam number ip
-- @return boolen, string
local function match_ipv4(self, ip)
    local ipv4s = self.ipv4
    local value = ipv4s[ip]
    if value ~= nil then
        return true, value
    end

    local ipv4_mask = self.ipv4_mask
    if self.ipv4_match_all_value ~= nil then
        return true, self.ipv4_match_all_value -- match any ip
    end

    for _, mask in ipairs(self.ipv4_mask_arr) do
        local valid_inet_addr = bit.rshift(ip, 32 - mask)           

        -- log.debug("ipv4 mask: ", mask,
        --          " valid inet: ", valid_inet_addr)

        value = ipv4_mask[mask][valid_inet_addr]
        if value ~= nil then
            return true, value
        end
    end

    return false, nil
end


-- @tparam number ip
-- @treturn boolen, string
local function match_ipv6(self, ip)
    local ipv6s = self.ipv6
    if self.ipv6_match_all_value ~= nil then
        return true, self.ipv6_match_all_value -- match any ip
    end

    for _, mask in ipairs(self.ipv6_mask_arr) do
        local node = ipv6s[mask]
        local inet_idxs = gen_ipv6_idxs(ip, mask)
        for _, inet in ipairs(inet_idxs) do
            if not node[inet] then
                break
            else
                node = node[inet]
                -- the case for the with_value is nil
                -- if node == true then
                --     return true
                -- end
                if type(node) == "number" then
                    -- fetch with the ipv6s_values_idx
                    local value = self.ipv6_values[node]
                    if value then
                        return true, value
                    end
                end
            end
        end
    end

    return false, nil
end

-- TODO: add a fore option
-- @tparam string ip_addr_org
-- @tparam string value
-- @treturn boolean [, string]
function _M.insert(self, ip_addr_org, value)
    if type(value) ~= "string" then
        return false, "value should a string"
    end

    if not is_valid_ip_or_cidr(ip_addr_org) then
        return false, "invalid ip address"
    end

    local ip_addr, ip_addr_mask = split_ip(ip_addr_org)

    -- for security
    if ip_addr_mask == 0 then
        return false, "add a cidr with mask 0 is not supported for security"

    end

    local inet_ipv4 = parse_ipv4(ip_addr)
    -- print(inet_ipv4)
    if inet_ipv4 then
        return insert_ipv4(self, inet_ipv4, ip_addr_mask, value)
    end

    local inets_ipv6 = parse_ipv6(ip_addr)
    if inets_ipv6 then
        return insert_ipv6(self, ip_addr, inets_ipv6, ip_addr_mask, value)
    end

    return false, "insert ip error"
end


-- @tparam string ip_addr_org
-- @treturn boolean [, string]
function _M.remove(self, ip_addr_org, force)
    if type(ip_addr_org) ~= "string" then
        return false, "ip_addr_org should a string"
    end

    if not is_valid_ip_or_cidr(ip_addr_org) then
        return false, "invalid ip address"
    end

    local ip_addr, ip_addr_mask = split_ip(ip_addr_org)
    local inet_ipv4 = parse_ipv4(ip_addr)

    if inet_ipv4 then
        return remove_ipv4(self, inet_ipv4, ip_addr_mask, force)
    end

    local inets_ipv6 = parse_ipv6(ip_addr)
    if inets_ipv6 then
        return remove_ipv6(self, ip_addr, inets_ipv6, ip_addr_mask, force)
    end

    return false, "remove ip error"

end

-- @tparam string ip
-- @return boolen, string [,string]
function _M.match(self, ip)
    local inet_ipv4 = parse_ipv4(ip)
    if inet_ipv4 then
        return match_ipv4(self, inet_ipv4)
    end

    local inets_ipv6 = parse_ipv6(ip)
    if not inets_ipv6 then
        return false, nil, "invalid ip address, not ipv4 and ipv6"
    end

    -- TODO: move to the match_ipv6 function
    local ipv6s = self.ipv6
    local value = ipv6s[ip]
    if value ~= nil then
        return true, value
    end

    return match_ipv6(self, inets_ipv6)
end

return _M