local rules       = require("configs").rules
local sites       = require("configs").sites
local log         = require("log")
local ngx         = ngx
local ngx_re      = ngx.re
local ngx_exit    = ngx.exit
local get_headers = ngx.req.get_headers

local NGX_OK                      = ngx.OK
local HTTP_GONE                   = ngx.HTTP_GONE
local HTTP_INTERNAL_SERVER_ERROR  = ngx.HTTP_INTERNAL_SERVER_ERROR

local _M = {
  _VERSION = 0.1
}

local function check_site_is_exist(ctx)
  if not sites.is_exist(ctx.var.host) then
    log.error("site ", ctx.var.host, " is not exist")
    return ctx.say_410()
  end
end


local function check_blacklist_ip(ctx)
  local client_ip = ctx.real_client_ip
  local rule = rules.blacklist_ip
  local matched, rule_id, err = rule:match(client_ip)
  if matched then
    log.block(ctx, "blacklist_ip")
    return ctx:say_block()
  end
  if err then
    log.error(err)
  end
end

local function check_blacklist_region(ctx)
  local client_ip = ctx.real_client_ip
  local mather = rules.blacklist_region
  -- if mather.match(ctx.site_id, client_ip) then
  if mather.match(ctx) then
    log.block(ctx, "blacklist_region")
    return ctx:say_block()
  end
end


local function check_whitelist_ip(ctx)
  local client_ip = ctx.real_client_ip
  local rule = rules.whitelist_ip
  local matched, rule_id, err = rule:match(client_ip)
  if matched then
    log.bypass(ctx, "whitelist_ip")
    return ngx_exit(NGX_OK)
  end
  if err then
    log.error(err)
  end
  return
end


local function check_whitelist_url(ctx)
  local host = ctx.var.host

  local path = ctx.request_path
  if not path then
    log.error("faild to fetching request url path")
    return ngx_exit(HTTP_INTERNAL_SERVER_ERROR)
  end

  local rule = rules.whitelist_url
  local rule_id = rule:match(host, path)
  if rule_id then
    log.bypass(ctx, "whitelist_url")
    return ngx_exit(NGX_OK)
  end
  return
end


local function check_ratelimit(ctx)
  local client_ip = ctx.real_client_ip
  local host = ctx.var.host
  local path = ctx.request_path
  if not path then
    log.error("faild to fetching request url path")
    return ngx_exit(HTTP_INTERNAL_SERVER_ERROR)
  end
  local rate_limit = rules.rate_limit
  local rule_id = rate_limit:url_match(host, path)
  if not rule_id then
    return
  end

  local count_key = host .. rule_id .. client_ip

  local delay, err = rate_limit:incomming(rule_id, count_key)
  if not delay then
    if err == "rejected" then
      log.block(ctx, "ratelimit")
      return ctx:say_block()
    end
    log.error("rate limit check failed")
    return ngx_exit(HTTP_INTERNAL_SERVER_ERROR)
  end
  return
end


function _M.access(_, ctx)
  check_site_is_exist(ctx)
  check_whitelist_ip(ctx)
  check_whitelist_url(ctx)
  check_blacklist_ip(ctx)
  check_blacklist_region(ctx)
  check_ratelimit(ctx)
end


return _M