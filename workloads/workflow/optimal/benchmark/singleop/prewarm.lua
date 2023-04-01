-- require "socket"
-- local JSON = require("JSON")
-- local UUID = require("uuid")
-- time = socket.gettime() * 1000
-- math.randomseed(time)
-- UUID.randomseed(time)

-- local function uuid()
--     return UUID():gsub('-', '')
-- end

request = function()
  local path = '/function/prewarm'
  local method = "POST"
  local headers = {}
  local body = '{' ..
    '"InstanceId": "",' ..
    '"CallerName": "",' ..
    '"Input":{}' ..
  '}'
  -- local body = JSON:encode(param)
  headers["Content-Type"] = "application/json"

  return wrk.format(method, path, headers, body)
end

function init(rand_seed)
    math.randomseed(rand_seed)
end
