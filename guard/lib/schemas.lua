local jsonschema    = require("resty.jsonschema")
local string_format = string.format
local table_insert  = table.insert
local table_concat  = table.concat
local pairs         = pairs

-- base
local client_ip_def, id_def, url_match_mode
do
    local ip_def
    do
        -- taken from https://github.com/apache/apisix/blob/master/apisix/schema_def.lua
        local _ipv4_seg = "([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"
        local _ipv4_def_buf = {}
        for i = 1, 4 do
            table_insert(_ipv4_def_buf, _ipv4_seg)
        end
        local _ipv4_def = table_concat(_ipv4_def_buf, [[\.]])
        -- There is false negative for ipv6/cidr. For instance, `:/8` will be valid.
        -- It is fine as the correct regex will be too complex.
        local _ipv6_def = "([a-fA-F0-9]{0,4}:){1,8}(:[a-fA-F0-9]{0,4}){0,8}"
                         .. "([a-fA-F0-9]{0,4})?"
        ip_def = {
            {title = "IPv4", type = "string", format = "ipv4"},
            {title = "IPv4/CIDR", type = "string", pattern = "^" .. _ipv4_def .. "/([12]?[0-9]|3[0-2])$"},
            {title = "IPv6", type = "string", format = "ipv6"},
            {title = "IPv6/CIDR", type = "string", pattern = "^" .. _ipv6_def .. "/[0-9]{1,3}$"},
        }
    end

    do
        client_ip_def = {
            description = "can be ipv4 or ipv6 or cidr",
            type = "string",
            anyOf = ip_def
        }
    end

    do
        id_def = {
            type = "string",
            pattern = "^[a-z0-9]{20}$"
        }        
    end

    do
        url_match_mode = {
            description = "the url match type: 1 means prefix match, 2 means exact match",
            type = "integer",
            enum = { 1, 2 },
            default = 1
        }
    end
end


-- delete
local delete_array_validator_def
do
    assert(id_def ~= nil, "id_def can not be empty")
    local delete_array_schema = {
        type = "array",
        items = id_def
    }
    delete_array_validator_def = jsonschema.generate_validator(delete_array_schema)
end


-- ip whitelist
local whitelist_ip_validator, whitelist_ip_schema_def
do
    whitelist_ip_schema_def = {
        type = "object",
        properties = {
            id  = id_def,
            ip  = client_ip_def
        },
        additionalProperties = false,
        required = {"id", "ip"}
    }

    local whitelist_ip_array_schema_def = {
        type = "array",
        items = whitelist_ip_schema_def
    }

    local whitelist_ip_array_validator_def  = jsonschema.generate_validator(whitelist_ip_array_schema_def)

    whitelist_ip_validator = {
        add = whitelist_ip_array_validator_def,
        del = delete_array_validator_def    
    }
end

-- ip blacklist
local blacklist_ip_validator, blacklist_ip_schema_def
do
    blacklist_ip_schema_def = {
        type = "object",
        properties = {
            id  = id_def,
            ip  = client_ip_def
        },
        additionalProperties = false,
        required = {"id", "ip"}
    }

    local blacklist_ip_array_schema_def = {
        type = "array",
        items = blacklist_ip_schema_def
    }

    local blacklist_ip_array_validator_def  = jsonschema.generate_validator(blacklist_ip_array_schema_def)

    blacklist_ip_validator = {
        add = blacklist_ip_array_validator_def,
        del = delete_array_validator_def    
    }
end

-- url whitelist
local whitelist_url_validator, whitelist_url_schema_def
do
    whitelist_url_schema_def = {
        type = "object",
        properties = {
            id  = id_def,
            site_id = id_def,
            host = { type = "string" },
            path = { type = "string" },
            match_mode = url_match_mode,
        },
        additionalProperties = false,
        required = { "id", "host", "path", "site_id", "match_mode" }
    }

    local whitelist_url_array_schema_def = {
        type = "array",
        items = whitelist_url_schema_def
    }

    local whitelist_url_array_validator_def = jsonschema.generate_validator(whitelist_url_array_schema_def)

    whitelist_url_validator = {
        add = whitelist_url_array_validator_def,
        del = delete_array_validator_def,
        update = whitelist_url_array_validator_def
    }    
end

-- region blacklist
local blacklist_region_validator, blacklist_region_schema_def
do
    blacklist_region_schema_def = {
        type = "object",
        properties = {
            site_id = id_def,
            countries = {
                type = "array",
                items = { type = "string"},
            },
            regions = {
                type = "array",
                items = { type = "string"},       
            },
            match_mode = {
                description = "match mode, 0 as blacklist, 1 as whitelist",
                type = "integer",
                enum = { 0, 1 },
                default = 0,
            },
        },
        additionalProperties = false,
        required = { "site_id", "countries", "regions", "match_mode" },    
    }

    local blacklist_region_array_schema_def = {
        type = "array",
        items = blacklist_region_schema_def,
    }

    local blacklist_region_array_validator_def = jsonschema.generate_validator(blacklist_region_array_schema_def)

    blacklist_region_validator = {
        add = blacklist_region_array_validator_def,
        del = delete_array_validator_def,
    }
end

-- rate limit
local rate_limit_validator, rate_limit_schema_def 
do
    rate_limit_schema_def = {
        type = "object",
        properties = {
            id = id_def,
            site_id = id_def,
            host = { type = "string" },
            path = { type = "string" },
            limit = { type = "integer", minimum = 1, default = 1 },
            window = { type = "integer",minimum = 1, default = 60 },
            match_mode = url_match_mode,
        },
        additionalProperties = false,
        required = { "id", "site_id", "host", "path", "limit", "window", "match_mode"}
    }
    local rate_limit_array_schema_def = {
        type = "array",
        items = rate_limit_schema_def
    }    
    local rate_limit_array_validator_def = jsonschema.generate_validator(rate_limit_array_schema_def)
    rate_limit_validator = {
        add = rate_limit_array_validator_def,
        update = rate_limit_array_validator_def,
        del = delete_array_validator_def,

    }
end


-- certificate
local certificate_validator, certificate_schema_def
do
    certificate_schema_def = {
        type = "object",
        properties = {
            id   = id_def,
            cert = "string",
            key  = "string",
            -- pem_cert_and_priv_key = "string"
        },
        additionalProperties = false,
        -- required = { "id", "pem_cert_and_priv_key" }
        required = { "id", "cert", "key" }
    }
    local certificate_array_schema_def = {
        type = "array",
        items = certificate_schema_def
    }
    local certificate_array_validator_def   = jsonschema.generate_validator(certificate_array_schema_def)
    certificate_validator = {
        add = certificate_array_validator_def,
        del = delete_array_validator_def,
        update = certificate_array_validator_def
    }
end


-- site
local site_validator, site_schema_def
do
    local site_config_def = {
    type = "object",
    properties = {
        is_https = {
            type = "integer",
            enum = { 0, 1 },
            default = 0,
        },
        cert_id = {
            type = "string",
            default = ""
        },
        origin_protocol = {
            type = "integer",
            enum = { 1, 2, 3 },
            default = 1
        },
        is_real_ip_from_header = {
            type = "integer",
            enum = { 0, 1 },
            default = 0,           
        },
        real_ip_header = {
            type = "string",
            default = "X-Forwarded-For"
        }
    },
    additionalProperties = false,
    required = { "is_https", "cert_id", "origin_protocol", "is_real_ip_from_header", "real_ip_header" }
}

    local site_origin_def = {
        type = "object",
        properties = {
            host = {
                type = "string"
            },
            http_port = {
                type = "integer",
                minimum = 1,
                maximum = 65535
            },
            https_port = {
                type = "integer",
                minimum = 1,
                maximum = 65535
            },        
            weight = {
                type = "integer",
                minimum = 0,
                maximum = 100,
                default = 100
            }
        },
        additionalProperties = false,
        required = { "host", "http_port", "https_port", "weight" }
    }


    site_schema_def = {
        type = "object",
        properties = {
            id = id_def,
            host = {
                type = "string"
            },
            config = site_config_def,
            origins = {
                type = "array",
                items = site_origin_def
            }
        },
        additionalProperties = false,
        required = { "id", "host", "config", "origins" }
    }

    local site_update_schema_def = {
        type = "object",
        properties = {
            id = id_def,
            host = {
                type = "string"
            },
            config = site_config_def,
            origins = {
                type = "array",
                items = site_origin_def
            }
        },
        additionalProperties = false,    
        anyOf = {
            { required = { "id", "host", "config" } },
            { required = { "id", "host", "origins" } }
        }
    }

    local site_array_schema = {
        type = "array",
        items = site_schema_def
    }

    local site_array_update_schema = {
        type = "array",
        items = site_update_schema_def
    }

    local site_array_validator_def          = jsonschema.generate_validator(site_array_schema)
    local site_array_udpate_validator_def   = jsonschema.generate_validator(site_array_update_schema)
    site_validator = {
        add = site_array_validator_def,
        del = delete_array_validator_def,
        update = site_array_udpate_validator_def
    }
end

-- log cfg
local log_cfg_validator, log_cfg_schema
do
    log_cfg_schema = {
        type = "object",
        properties = {
            is_sample_log_upload = {
                type = "integer",
                enum = { 0, 1 },
                default = 0,           
            },
            sample_log_upload_api        = { type = "string" },
            sample_log_upload_api_token  = { type = "string" }
        },
        additionalProperties = false,
        required = { "is_sample_log_upload", "sample_log_upload_api", "sample_log_upload_api_token"}
    }
    log_cfg_validator = {
        update = jsonschema.generate_validator(log_cfg_schema)
    }
end


-- full configuration
local full_cfg_validator
do
    local full_cfg_schema = {
        type = "object",
        properties = {
            log             = log_cfg_schema,
            sites           = { type = "array", items = site_schema_def },
            certificates    = { type = "array", items = certificate_schema_def },
            rules = {
                type = "object",
                properties = {
                    blacklist_ip     = { type = "array", items = blacklist_ip_schema_def },
                    blacklist_region = { type = "array", items = blacklist_region_schema_def },
                    whitelist_ip     = { type = "array", items = whitelist_ip_schema_def },
                    whitelist_url    = { type = "array", items = whitelist_url_schema_def },
                    rate_limit       = { type = "array", items = rate_limit_schema_def },
                },
                additionalProperties = false,
                required = { "blacklist_ip", "blacklist_region", "whitelist_ip", "whitelist_url", "rate_limit" }
            },
        },
        additionalProperties = false,
        required = { "log", "sites", "certificates", "rules"}

    }

    local full_sync_validator_def = jsonschema.generate_validator(full_cfg_schema)

    full_validator = {
        full_sync = full_sync_validator_def
    }
end


local validator = {
    this             = full_validator,
    sites            = site_validator,
    sample_log       = log_cfg_validator,
    certificates     = certificate_validator,
    blacklist_ip     = blacklist_ip_validator,
    whitelist_ip     = whitelist_ip_validator,
    whitelist_url    = whitelist_url_validator,
    rate_limit       = rate_limit_validator,
    blacklist_region = blacklist_region_validator,
}

do
    local not_support_operation = {
        __index = function(_, operation_type)
            return function(_)
                return false, string_format("unsupported operation type '%s'", operation_type)
            end
        end
    }

    local not_support_config_type = {
        __index = function(_, config_type)
            return setmetatable({},{
                __index = function(t, k)
                    return function(_)
                        return false, string_format("unsupported config type '%s'", config_type)
                    end
                end
            })
        end
    }

    setmetatable(validator, not_support_config_type)

    for _,v in pairs(validator) do
        setmetatable(v, not_support_operation)
    end
end

return {
    validator = validator
}