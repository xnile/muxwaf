local router  = require("router")
local metrics = require("apis.metrics")
local configs = require("apis.configs")


local r = router.new("/api")
r:POST("/sites", configs.sites.add)
r:PUT("/sites", configs.sites.update)
r:DELETE("/sites", configs.sites.del)

r:POST("/certificates", configs.certificates.add)
r:PUT("/certificates", configs.certificates.update)
r:DELETE("/certificates", configs.certificates.del)

r:POST("/blacklist/ip", configs.blacklist_ip.add)
r:DELETE("/blacklist/ip", configs.blacklist_ip.del)

r:PUT("/blacklist/region", configs.blacklist_region.add)
r:DELETE("/blacklist/region", configs.blacklist_region.del)

r:POST("/whitelist/ip", configs.whitelist_ip.add)
r:DELETE("/whitelist/ip", configs.whitelist_ip.del)

r:POST("/whitelist/url", configs.whitelist_url.add)
r:PUT("/whitelist/url", configs.whitelist_url.update)
r:DELETE("/whitelist/url", configs.whitelist_url.del)

r:POST("/rate-limit", configs.rate_limit.add)
r:PUT("/rate-limit", configs.rate_limit.update)
r:DELETE("/rate-limit", configs.rate_limit.del)

r:PUT("/sys/configs/sample_log_upload",configs.sample_log.update)
r:GET("/sys/configs", configs.this.show)
r:POST("/sys/configs", configs.this.full_sync)
r:GET("/sys/metrics", metrics.get)
r:DELETE("/sys/configs/rules", configs.rules.reset)
r:DELETE("/sys/configs", configs.this.reset)


return r