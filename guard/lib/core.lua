local rules       = require("configs").rules
local sites       = require("configs").sites
local constants   = require("constants")
local sample_log  = require("sample_log")
local log         = require("log")
local ngx         = ngx
local ngx_re      = ngx.re
local ngx_exit    = ngx.exit
local get_headers = ngx.req.get_headers
local tostring    = tostring

local NGX_OK                      = ngx.OK
local HTTP_GONE                   = ngx.HTTP_GONE
local HTTP_INTERNAL_SERVER_ERROR  = ngx.HTTP_INTERNAL_SERVER_ERROR
local RULE_TYPE                   = constants.RULE_TYPE
-- local DEFAULT_API_LISTEN_PORT     = constants.DEFAULT_API_LISTEN_PORT

local _M = {
  _VERSION = 0.1
}

-- local function bypass_apis(ctx)
--   if ctx.server_port == tostring(DEFAULT_API_LISTEN_PORT) then
--     -- log.debug("allow apis to pass through")
--     return ngx_exit(NGX_OK)
--   end
-- end


local function check_site_is_exist(ctx)
  if ctx.site_id == "" then
    log.warn("The site '", ctx.host, "' is not exist")
    return ctx.say_410()
  end
end


local function check_blacklist_ip(ctx)
  local client_ip = ctx.real_client_ip
  local rule = rules.blacklist_ip
  local matched, rule_id, err = rule:match(client_ip)
  if matched then
    sample_log.block(ctx, RULE_TYPE.BLACKLIST_IP, rule_id)
    return ctx:say_block()
  end
  if err then
    log.warn(err)
  end
end

local function check_blacklist_region(ctx)
  local client_ip = ctx.real_client_ip
  local mather = rules.blacklist_region
  if mather.deny(ctx) then
    sample_log.block(ctx, RULE_TYPE.BLACKLIST_REGION, nil)
    return ctx:say_block()
  end
end


local function check_whitelist_ip(ctx)
  local client_ip = ctx.real_client_ip
  local rule = rules.whitelist_ip
  local matched, rule_id, err = rule:match(client_ip)
  if matched then
    sample_log.bypass(ctx, RULE_TYPE.WHITELIST_IP, rule_id)
    return ngx_exit(NGX_OK)
  end
  if err then
    log.warn(err)
  end
  return
end


local function check_whitelist_url(ctx)
  local host = ctx.host

  local path = ctx.request_path
  if not path then
    log.error("faild to fetching request url path")
    return ngx_exit(HTTP_INTERNAL_SERVER_ERROR)
  end

  local rule = rules.whitelist_url
  local rule_id = rule:match(host, path)
  if rule_id then
    sample_log.bypass(ctx, RULE_TYPE.WHITELIST_URL, rule_id)
    return ngx_exit(NGX_OK)
  end
  return
end


local function check_ratelimit(ctx)
  local client_ip = ctx.real_client_ip
  local host = ctx.host
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
      sample_log.block(ctx, RULE_TYPE.RATELIMIT, rule_id)
      return ctx:say_block()
    end
    log.error("rate limit check failed")
    return ngx_exit(HTTP_INTERNAL_SERVER_ERROR)
  end
  return
end


function _M.access_phase(ctx)
  -- bypass_apis(ctx)
  check_site_is_exist(ctx)
  check_whitelist_ip(ctx)
  check_whitelist_url(ctx)
  check_blacklist_ip(ctx)
  check_blacklist_region(ctx)
  check_ratelimit(ctx)
end


return _M