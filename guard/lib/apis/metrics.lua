local metrics = require("metrics")

return {
    get = function(c)
        return c.say_json(metrics.show())
    end
}