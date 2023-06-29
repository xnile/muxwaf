local events   = require("events")
local schemas  = require("schemas")
local configs  = require("configs")
local json     = require("utils.json")
local metrics  = require("metrics")
local setmetatable = setmetatable


return setmetatable({}, {
    __index = function(_, configType)
        return setmetatable({}, {
            __index = function(_, operation)
                return function(c)
                    local event = {
                        configType = configType,
                        operation = operation,
                        data  = {},
                    }

                    if operation == "reset" then
                        events:send(c.encode(event))
                        return c.say_ok()
                    elseif operation == "show" then
                        c.say_json(configs.get_raw())
                    else
                        local data, err = c.get_and_decode_body_data()
                        if not data then
                            return c.say_err(500, "invalid json data: " .. err)
                        end

                        local validator = schemas.validator[configType][operation]
                        local ok, err = validator(data)
                        if not ok then
                            return c.say_err(500, err)
                        end

                        event.data = data
                        events:send(c.encode(event))
                        metrics:incr_config_updates()
                        return c.say_ok()
                    end
                end
            end
        })
    end
})
