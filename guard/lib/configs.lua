local require        = require
local ditcs          = require("constants").DICTS
local cjson          = require("cjson.safe")
local log            = require("log")
local ngx_timer_at   = ngx.timer.at
local io_open        = io.open
local io_close       = io.close
local pairs          = pairs
local setmetatable   = setmetatable
local tostring       = tostring
local table_new      = table.new
local ngx_shared     = ngx.shared
local ngx_cfg_prefix = ngx.config.prefix

local CONFIG_FILE = ngx_cfg_prefix() .. "config.json"

local _M = {
  _VERSION = 0.1
}

local rules = {
    blacklist_ip     = require("core.blacklist.ip"),
    blacklist_region = require("core.blacklist.region"),
    whitelist_ip     = require("core.whitelist.ip"),
    whitelist_url    = require("core.whitelist.url"),
    rate_limit       = require("core.ratelimit"),    
}

setmetatable(rules, {
    __metatable = false,
    __index = {
        -- add = function(_, cfg)
        --     for k, v in pairs(rules) do
        --         v:add(cfg[k])  --  cfg[k] is array like table
        --     end
        -- end

        full_sync = function(_, cfg)
            for k, v in pairs(rules) do
                v:full_sync(cfg[k])  --  cfg[k] is array like table
            end
        end

        ,reset = function(_)
            for _, v in pairs(rules) do
                v:reset()
            end
        end

        ,get_raw = function(_)
            local cfg = {}
            for k, v in pairs(rules) do
                cfg[k] = v:get_raw()
            end
            return cfg
        end
    }
})


local configs = {
    certificates = require("certificates"),
    sites = require("sites"),
    log = log,
    rules = rules
}

setmetatable(configs, {
    __metatable = false,
    __index = rules
})


local this = {}
setmetatable(this, {
    __index = {
        reset = function()
            for _, v in pairs(configs) do
                v:reset()
            end
        end

        ,full_sync = function(_, cfg)
            for k, v in pairs(cfg) do
                if configs[k] then
                    configs[k]:full_sync(cfg[k])
                end
            end
        end

        ,get_cnt = function()
            local cfg = {}
            for k, c in pairs(configs) do
                cfg[k] = c:get_raw()
            end
            return cfg    
        end
    }
})

local config_types = {
    this = this
}
setmetatable(config_types, {
    __index = configs
})




-- local function get_cnt_configs(_)
--     local cfg = {}
--     for k, c in pairs(configs) do
--         cfg[k] = c:get_raw()
--     end
--     return cfg
-- end

local async_save_config
do
    local function save()
        local config = this.get_cnt()
        local fd, err = io_open(CONFIG_FILE, "w")
        if not fd then
            log.error("failed to open config file: ", tostring(err))
            return
        end
        
        fd:write(cjson.encode(config))
        fd:flush()
        io_close(fd)        
    end

    async_save_config = function()
        local _, err = ngx_timer_at(10, save) -- TODO: time can be customized
        if err then
            log.error("failed to setting up timer for save config: ", tostring(err))
            return
        end
    end
end

-- local function sync_cfg_from_cfg_file(cfg)

--     -- Exit if file is empty
--     if not cfg then return end

--     for k, v in pairs(cfg) do
--         if configs[k] then
--             configs[k]:full_sync(cfg[k])
--         end
--     end    

-- end

-- full configs sync
-- local function apply_full_sync(cfg)
--     if not cfg then return end

--     for k, v in pairs(cfg) do
--         if configs[k] then
--             configs[k]:full_sync(cfg[k])
--         end
--     end

-- end

-- local function apply_sync(target, operation, cfg)
--   if operation == "add" then
--     target:add(cfg)
--   elseif operation == "del" then
--     target:del(cfg)
--   elseif operation == "update" then
--     target:update(cfg)
--   elseif operation == "reset" then
--     target:reset()
--   elseif operation == "sync" then
--     -- full configuration
--     apply_full_sync(cfg)
--   else
--     log.warn(operation .. " operation is not supported")
--     return
--   end

--   -- trigger save config to file
--   async_save_config()
-- end

-- local function apply_sync(target, operation, cfg)
--     if operation == "sync" then
--         -- full configuration
--         apply_full_sync(cfg)
--     else
--         target[operation](nil,cfg)
--     end

--       -- trigger save config to file
--     async_save_config()
-- end


function _M.init()
    local fd, err = io_open(CONFIG_FILE, "r")
    if not fd then
        fd = io_open(CONFIG_FILE, "w")
        assert(fd, "Failed to create config file: " .. CONFIG_FILE)
        fd:write("{}")
        fd:flush()
        fd:close()
        return
    end

    local data  = fd:read("*a")
    fd:close()

    cfg = cjson.decode(data)
    if not cfg then return end
    this.full_sync(nil, cfg)

end


function _M.sync(_, configType, operation, cfg)
    if not cfg then
        log.warn("failed to sync configs: ","config is empty")
        return
    end
    local target = config_types[configType]
    if not target then
        log.warn(configType.." config type is not supported")
        return
    end

    if not target[operation] then
        log.warn(operation.." operation is not supported")
        return
    end

    target[operation](nil, cfg)
    async_save_config()
end


function _M.get_raw(_)
    return this.get_cnt()
end

local _mt = {
    sites = configs.sites,
    certificates = configs.certificates,
    rules = rules
}

setmetatable(_M, {
    __index = _mt
})

return _M