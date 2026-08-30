package oidcsessionredis

const gateScript = `
local raw = redis.call('GET', KEYS[1])
if not raw then return false end
local record = cjson.decode(raw)
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
if tonumber(record.access_expires_at) > now + window then
  record.last_refresh_at = now
  local expiry = math.min(now + tonumber(ARGV[3]), tonumber(record.absolute_expires_at))
  local updated = cjson.encode(record)
  redis.call('SET', KEYS[1], updated, 'PXAT', expiry)
  return {1, updated}
end
if record.lease_owner and record.lease_owner ~= '' and tonumber(record.lease_until or 0) > now then
  return {2}
end
record.lease_owner = ARGV[4]
record.lease_until = now + tonumber(ARGV[5])
local leased = cjson.encode(record)
redis.call('SET', KEYS[1], leased, 'KEEPTTL')
return {3, leased}
`

const commitScript = `
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
local record = cjson.decode(raw)
if record.lease_owner ~= ARGV[1] then return 0 end
record.payload = ARGV[2]
record.access_expires_at = tonumber(ARGV[3])
record.last_refresh_at = tonumber(ARGV[4])
record.lease_owner = ''
record.lease_until = 0
local expiry = math.min(tonumber(ARGV[4]) + tonumber(ARGV[5]), tonumber(record.absolute_expires_at))
local updated = cjson.encode(record)
redis.call('SET', KEYS[1], updated, 'PXAT', expiry)
return 1
`

const releaseScript = `
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
local record = cjson.decode(raw)
if record.lease_owner ~= ARGV[1] then return 0 end
record.lease_owner = ''
record.lease_until = 0
redis.call('SET', KEYS[1], cjson.encode(record), 'KEEPTTL')
return 1
`
