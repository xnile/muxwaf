local resty_counter = require("resty.counter")


local _M = {}
local shdict_name = "limit"
local sync_interval = 1


local counter



function _M.incomming(key, limit, window)

end


function _M.init_worker()
    counter = resty_counter.new(shdict_name, sync_interval)
end


local function inc(key, step)



function _M.log_phase()
    local key = "xxx"
    -- local remaining, err = dict:incr(key, -1, limit, window)
     -- local delay, err = rate_limit:incomming(rule_id, count_key)
    counter:incr(key, -1)
end