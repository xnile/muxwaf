local metrics = require("metrics")

return {
    get = function(c)
        return metrics.collect(c)
    end
}