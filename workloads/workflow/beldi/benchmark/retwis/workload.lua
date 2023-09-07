-- require "socket"
-- local JSON = require("JSON")
-- local UUID = require("uuid")
-- time = socket.gettime() * 1000
-- UUID.randomseed(time)
-- math.randomseed(socket.gettime() * 1000)
math.random();
math.random();
math.random()

-- local function uuid()
--     return UUID():gsub('-', '')
-- end

-- local gatewayPath = os.getenv("ENDPOINT")

local function get_user()
    local id = math.random(0, 10000)
    local user_name = "user" .. tostring(id)
    local pass_word = "pwd" .. tostring(id)
    -- for i = 0, 9, 1 do
    --     pass_word = pass_word .. tostring(id)
    -- end
    return user_name, pass_word
end

local function login()
    local user_name, password = get_user()
    local method = "POST"
    local path = '/asyncFunction/login'
    -- local param = {
    --     InstanceId = uuid(),
    --     CallerName = "",
    --     Async = true,
    --     Input = {
    --         Function = "user",
    --         Input = {
    --             Username = user_name,
    --             Password = password
    --         }
    --     }
    -- }
    local body = '{' ..
    --   '"InstanceId": "' .. uuid() .. '",' ..
      '"InstanceId": "",' ..
      '"CallerName": "",' .. '"Async": true,' ..
      '"Input":{' ..
        '"Function":"Login",' ..
        '"Input": {' ..
          '"Username": "' .. user_name .. '",' ..
          '"Password": "' .. password .. '"' ..
        '}' ..
      '}' ..
    '}'
    local headers = {}
    headers["Content-Type"] = "application/json"

    return wrk.format(method, path, headers, body)
end

local function profile()
  local user_name, password = get_user()
  local method = "POST"
  local path = '/asyncFunction/profile'
  -- local param = {
  --     InstanceId = uuid(),
  --     CallerName = "",
  --     Async = true,
  --     Input = {
  --         Function = "user",
  --         Input = {
  --             Username = user_name,
  --             Password = password
  --         }
  --     }
  -- }
  local body = '{' ..
  --   '"InstanceId": "' .. uuid() .. '",' ..
    '"InstanceId": "",' ..
    '"CallerName": "",' .. '"Async": true,' ..
    '"Input":{' ..
      '"Function":"Profile",' ..
      '"Input": {' ..
        '"Username": "' .. user_name .. '"' ..
      '}' ..
    '}' ..
  '}'
  local headers = {}
  headers["Content-Type"] = "application/json"

  return wrk.format(method, path, headers, body)
end

local function post()
  local user_name, password = get_user()
  local method = "POST"
  local path = '/asyncFunction/post'
  -- local param = {
  --     InstanceId = uuid(),
  --     CallerName = "",
  --     Async = true,
  --     Input = {
  --         Function = "user",
  --         Input = {
  --             Username = user_name,
  --             Password = password
  --         }
  --     }
  -- }
  local body = '{' ..
  --   '"InstanceId": "' .. uuid() .. '",' ..
    '"InstanceId": "",' ..
    '"CallerName": "",' .. '"Async": true,' ..
    '"Input":{' ..
      '"Function":"Post",' ..
      '"Input": {' ..
        '"Username": "' .. user_name .. '",' ..
        '"Content": ""' ..
      '}' ..
    '}' ..
  '}'
  local headers = {}
  headers["Content-Type"] = "application/json"

  return wrk.format(method, path, headers, body)
end

local function timeline()
  local user_name, password = get_user()
  local method = "POST"
  local path = '/asyncFunction/timeline'
  -- local param = {
  --     InstanceId = uuid(),
  --     CallerName = "",
  --     Async = true,
  --     Input = {
  --         Function = "user",
  --         Input = {
  --             Username = user_name,
  --             Password = password
  --         }
  --     }
  -- }
  local body = '{' ..
  --   '"InstanceId": "' .. uuid() .. '",' ..
    '"InstanceId": "",' ..
    '"CallerName": "",' .. '"Async": true,' ..
    '"Input":{' ..
      '"Function":"Timeline",' ..
      '"Input": {' ..
        '"Username": "' .. user_name .. '"' ..
      '}' ..
    '}' ..
  '}'
  local headers = {}
  headers["Content-Type"] = "application/json"

  return wrk.format(method, path, headers, body)
end

request = function()
    -- cur_time = math.floor(socket.gettime())
    local login_ratio = 0.05
    local profile_ratio = 0.05
    local timeline_ratio = 0.8
    local post_ratio = 0.1

    --return search_hotel()
    --return recommend()
    --return user_login()
    --return reserve_all()
    local coin = math.random()
    if coin < login_ratio then
        return login()
    elseif coin < login_ratio + profile_ratio then
        return profile()
    elseif coin < login_ratio + profile_ratio + timeline_ratio then
        return timeline()
    else
        return post()
    end
end

function init(rand_seed)
    math.randomseed(rand_seed)
end
