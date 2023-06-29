
local _M = { 
  _VERSION = 0.1
}


function _M.add_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to add origin of the site '", host, "', origin already exists")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
end


function _M.del_origins(_, items)
    for _, host in ipairs(items) do
        if not origins[host] then
            log.warn("Failed to remove the origin server of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = nil
        balancers[host] = nil
        upstream_servers[host] = nil
        ::continue::
    end
end


function _M.update_origins(_, items)
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins

        if origins[host] then
            log.warn("failed to update origin of the site '", host, "', the site does not exist")
            goto continue
        end

        site_origins[host] = table_clone(origins)
        ::continue::
    end
end


function _M.full_sync_origins(_, items)
    local new_origins = {}
    for _, item in ipairs(items) do
        local host, origins = item.host, item.origins
        new_origins[host] = table_clone(origins)
    end
    site_origins = new_origins
end
