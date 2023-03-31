local require      = require
local _            = require("cjson.safe").encode_empty_table_as_object(false)
-- local _            = require("utils.table.deepcopy")
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

local ctx
local IPIPDB_FILE  = ngx.config.prefix() .. "ipdb/ipipfree.ipdb"
local ipipdb       = ipdb_parser:new(IPIPDB_FILE)

function _M.get_ipdb()
  return {
    ipip = ipipdb
  }
end

function _M.init_phase()
  do
    for _, dict in pairs(constants.DICTS) do
      assert(ngx.shared[dict], "the lua_shared_dict '" .. dict .. "' undefined")
    end    
  end

  -- require("log").init()
  sample_log.init()
end

function _M.init_worker_phase()
  require("tasks").run()
  require("configs").init()
end

function _M.exit_worker_phase()
  -- local sample_log = require("log")
  sample_log.worker_exit()
end


local mod_ctx -- lazy load
function _M.access_phase()
  if not mod_ctx then
    mod_ctx = require('ctx')
  end
  ctx = mod_ctx.new()
  
  local core = require("core")
  core:access(ctx)
end

function _M.balance_phase()
  local balancer = require("balancer")
  balancer:balance(ctx)
end

 -- before access phase, ctx not ready
function _M.ssl_certificate_phase()
  local ssl = require("ssl")
  ssl.certificate()
end

function _M.log_phase()
  tablepool.release("pool_ctx", ctx)
  metrics.incr_resp_sts_code()
end

function _M.api_serve()
  -- ctx = require("ctx").new()
  apis:start(ctx)
end

function _M.say_500()
  ngx.header["Content-Type"] = "text/html"
  ngx.say(page_500)
end

return _M