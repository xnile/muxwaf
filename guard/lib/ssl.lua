local require       = require
local ssl           = require("ngx.ssl")
local log           = require("log")
local constants     = require("constants")
local sites         = require("configs").sites
local certificates  = require("configs").certificates
local lrucache      = require("resty.lrucache")
local ngx           = ngx
local ngx_exit      = ngx.exit
local error         = error
local string_format = string.format
local NGX_ERROR     = ngx.ERROR

local _M = {
    _VERSION = 0.1
}

local CACHE_TTL  = 300  -- in seconds, time of certificate take effect after updated
local CACHE_SIZE = 100
local DEFAULT_CERT, DEFAULT_PRIV_KEY

local cache, err = lrucache.new(CACHE_SIZE * 2)
if not cache then
    error("failed to create the cache: " .. (err or "unknown"), 2)
end

local function parse_pem_cert_and_priv_key(pem_cert)
    local cert, err = ssl.parse_pem_cert(pem_cert)
    if not cert then
        return nil, nil, "failed to parse PEM cert: " .. err
    end

    local pkey, err = ssl.parse_pem_priv_key(pem_cert)
    if not pkey then
        return nil, nil, "failed to parse PEM key: " .. err
    end

    return cert, pkey, nil
end

DEFAULT_CERT, DEFAULT_PRIV_KEY = parse_pem_cert_and_priv_key(constants.DEFAULT_CERT)

local function set_cert_and_priv_key(cert, pkey)
    local ok, err = ssl.set_cert(cert)
    if not ok then
        log.error("failed to set certificate: ", err)
        return ngx_exit(NGX_ERROR)
    end

    local ok, err = ssl.set_priv_key(pkey)
    if not ok then
        log.error("failed to set private key : ", err)
        return ngx_exit(NGX_ERROR)
    end
end

local function get_cert_and_priv_key_by_cert_id(cert_id)
    local cert_ckey, pkey_ckey = "cert:" .. cert_id, "pkey:" .. cert_id

    local cert, pkey = cache:get(cert_ckey), cache:get(pkey_ckey)
    if not cert or not pkey then
        local pem_cert_and_pkey, err = certificates.get(cert_id)
        if not pem_cert_and_pkey then
            return nil, nil, err
        end

        local cert, pkey, err = parse_pem_cert_and_priv_key(pem_cert_and_pkey)
        if not cert or not pkey then
            return nil, nil, err
        end

        cache:set(cert_ckey, cert, CACHE_TTL)
        cache:set(pkey_ckey, pkey, CACHE_TTL)
        return cert, pkey, nil
    end

    return cert, pkey, nil
end

function  _M.certificate(ctx)
    local server_name, err = ssl.server_name()
    if err then
        log.error("an error was encountered while obtainning Server Name: ", err)
    end

    if not server_name then
        log.warn("the client does not support SNI, falling back to default certificate")
        return set_cert_and_priv_key(DEFAULT_CERT, DEFAULT_PRIV_KEY)
    end

    if not sites.is_exist(server_name) then
        log.info(string_format("the site %s does not exist", server_name))
        return set_cert_and_priv_key(DEFAULT_CERT, DEFAULT_PRIV_KEY)
    end

    if not sites.is_enable_https(server_name) then
        log.info(string_format("the site %s https not enabled", server_name))
        return set_cert_and_priv_key(DEFAULT_CERT, DEFAULT_PRIV_KEY)
    end

    local cert_id = sites.get_site_cert_id(server_name)
    if cert_id == nil or cert_id == "" then
        log.info("the cert id is empty")
        return set_cert_and_priv_key(DEFAULT_CERT, DEFAULT_PRIV_KEY)
    end

    local cert, pkey, err = get_cert_and_priv_key_by_cert_id(cert_id)
    if not cert or not pkey then
        log.error(string_format("failed to get certificate and private key by id %s : %s",cert_id, err))
        return set_cert_and_priv_key(DEFAULT_CERT, DEFAULT_PRIV_KEY)
    end

    local ok, err = ssl.clear_certs()
        if not ok then
        log.error("failed to clear existing certificates: ", err)
        return ngx_exit(NGX_ERROR)
    end

    -- local err = set_cert_and_priv_key(cert, pkey)
    -- if err then
    --     log.error("failed to set certificate and private key : ", err)
    --     return ngx_exit(NGX_ERROR)
    -- end
    
    return set_cert_and_priv_key(cert, pkey)
end


return _M