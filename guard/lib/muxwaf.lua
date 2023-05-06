local require      = require
local _            = require("cjson.safe").encode_empty_table_as_object(false)
local ctxdump      = require("resty.ctxdump")
local tablepool    = require("resty.tablepool")
local ipdb_parser  = require("resty.ipdb.city")
local constants    = require("constants")
local page_500     = require("page.500")
local sample_log   = require("sample_log")
local apis         = require("apis")
local time         = require("time")
local metrics      = require("metrics")
local setmetatable = setmetatable
local tab_new      = table.new
local assert       = assert
local pairs        = pairs
local ngx          = ngx

local _M = {
  _VERSION = 0.1
}

local IPIPDB_FILE  = ngx.config.prefix() .. "ipdb/ipipfree.ipdb"
local ipipdb       = ipdb_parser:new(IPIPDB_FILE)

function _M.get_ipdb()
  return {
    ipip = ipipdb
  }
end

function _M.init_phase()
  for _, shdict_name in pairs(constants.DICTS) do
    assert(ngx.shared[shdict_name], "shared dict \"" .. (shdict_name or "nil") .. "\" not defined")
  end

  -- require("log").init()
  sample_log.init()
end

function _M.init_worker_phase()
  require("tasks").run()
  require("configs").init()
  require("metrics").init_worker()
end

function _M.exit_worker_phase()
  sample_log.worker_exit()
end


local mod_ctx -- lazy load
function _M.access_phase()
  if not mod_ctx then
    mod_ctx = require('ctx')
  end
  ngx.ctx.waf_ctx = mod_ctx.new()
  ngx.var.ctx_ref = ctxdump.stash_ngx_ctx()

  local core = require("core")
  core:access(ngx.ctx.waf_ctx)
end

function _M.balance_phase()
  local balancer = require("balancer")
  balancer.balance(ngx.ctx.waf_ctx)
end

 -- before access phase, ctx not ready
function _M.ssl_certificate_phase()
  local ssl = require("ssl")
  ssl.certificate()
end

function _M.log_phase()
  local ref = ngx.var.ctx_ref
  if ref ~= '' then
    local stash_ctx = ctxdump.apply_ngx_ctx(ref)
    ngx.var.ctx_ref = ''
    if not ngx.ctx.waf_ctx then
      ngx.ctx = stash_ctx
    end
  end

  local ctx = ngx.ctx.waf_ctx
  if not ctx then return end

  sample_log.log_phase(ctx)
  metrics.log_phase(ctx)
  tablepool.release("pool_ctx", ctx)
end

function _M.api_serve()
  local ctx = ngx.ctx.waf_ctx
  apis:start(ngx.ctx.waf_ctx)
end

function _M.say_500()
  ngx.header["Content-Type"] = "text/html"
  ngx.say(page_500)
end

return _M