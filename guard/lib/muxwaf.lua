local require      = require
local _            = require("cjson.safe").encode_empty_table_as_object(false)
local ctxdump      = require("resty.ctxdump")
local tablepool    = require("resty.tablepool")
local geo          = require("geo")
local constants    = require("constants")
local page_500     = require("page.500")
local page_410     = require("page.410")
local sample_log   = require("sample_log")
local apis         = require("apis")
local time         = require("time")
local metrics      = require("metrics")
local tasks        = require("tasks")
local balancer     = require("balancer")
local configs      = require("configs")
local core         = require("core")
local sites        = require("sites")
local setmetatable = setmetatable
local table_concat = table.concat
local assert       = assert
local pairs        = pairs
local ngx          = ngx

local _M = {
  _VERSION = 0.1
}

local geo_ip_searcher
do
  geo_ip_searcher = geo.new("ipip")
  -- geo_ip_searcher = geo.new("xdb")
end

function _M.get_ip_geo_searcher()
  return geo_ip_searcher
end

function _M.init_phase()
  for _, shdict_name in pairs(constants.DICTS) do
    assert(ngx.shared[shdict_name], "shared dict \"" .. (shdict_name or "nil") .. "\" not defined")
  end

  -- require("log").init()
  sample_log.init()
end

function _M.init_worker_phase()
  tasks.init_worker()
  configs.init_worker()
  metrics.init_worker()
  balancer.init_worker()
end

function _M.exit_worker_phase()
  sample_log.worker_exit()
end


local mod_ctx -- lazy load
local function ctx_init()
  if not mod_ctx then
    mod_ctx = require('ctx')
  end

  ngx.ctx.waf_ctx = mod_ctx.new()
  ngx.var.ctx_ref = ctxdump.stash_ngx_ctx()
end

function _M.rewrite_phase()
  local host = ngx.var.host
  if not sites.is_exist(host) then
    ngx.header["Content-Type"] = "text/html"
    ngx.status = 410
    ngx.say(page_410)
    ngx.exit(ngx.status)
    return
  end

  local request_uri = ngx.var.request_uri
  local scheme  = ngx.var.scheme
  if scheme == "http" and sites.is_force_https(host) then
    return ngx.redirect(table_concat({"https://", host, request_uri}), 302)
  end
  ctx_init()
end


function _M.access_phase()
  -- ctx_init()

  local ctx = ngx.ctx.waf_ctx
  core.access_phase(ctx)
  balancer.access_phase(ctx)
end

function _M.balance_phase()
  -- local balancer = require("balancer")
  local ctx = ngx.ctx.waf_ctx
  balancer.balance(ctx)
end

 -- before access phase, ctx not ready
function _M.ssl_certificate_phase()
  local ssl = require("ssl")
  ssl.certificate()
end

local function stash_ctx()
  local ref = ngx.var.ctx_ref
  if ref ~= '' then
    local stash_ctx = ctxdump.apply_ngx_ctx(ref)
    ngx.var.ctx_ref = ''
    if not ngx.ctx.waf_ctx then
      ngx.ctx = stash_ctx
    end
  end
end

function _M.log_phase()
  stash_ctx()

  local ctx = ngx.ctx.waf_ctx
  if not ctx then return end

  sample_log.log_phase(ctx)
  metrics.log_phase(ctx)
  tablepool.release("pool_ctx", ctx)
end

function _M.api_serve()
  apis:start()
end

function _M.say_500()
  ngx.header["Content-Type"] = "text/html"
  ngx.say(page_500)
end

return _M