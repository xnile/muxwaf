local dns_client

local setup_client = function()
  if not dns_client then
    dns_client = require "resty.dns.client"
  end

  local opts = {
    hosts = nil,                                         -- defaults to /etc/hosts
    resolvConf = nil,                                    -- defaults to system resolv.conf
    nameservers = nil,                                   -- provided list or taken from resolv.conf
    enable_ipv6 = false,                                 -- allow for ipv6 nameserver addresses
    retrans = nil,                                       -- taken from system resolv.conf; attempts
    timeout = nil,                                       -- taken from system resolv.conf; timeout
    validTtl = nil,                                      -- ttl in seconds overriding ttl of valid records,  only entries from hosts 
    badTtl = 1,                                          -- ttl in seconds for dns error responses (except 3 - name error)
    emptyTtl = 30,                                       -- ttl in seconds for empty and "(3) name error" dns responses
    staleTtl = 4,                                        -- ttl in seconds for records once they become stale
    cacheSize = 10000,                                   -- maximum number of records cached in memory
    order = { "last", "SRV", "A", "AAAA", "CNAME" },     -- order of trying record types
    noSynchronisation = false,                           -- Disables synchronization between queries, resulting in each lookup for the
                                                         -- -- same name being executed in it's own query to the nameservers. The default
                                                         -- -- (`false`) will synchronize multiple queries for the same name to a single
                                                         -- -- query to the nameserver.
  }

  assert(dns_client.init(opts))

  return dns_client
end

return setup_client