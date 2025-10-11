#!/usr/bin/env bash
set -euo pipefail

# Simple smoke test for the API running on localhost.
# Performs: healthz -> create user -> list -> get -> update -> delete -> verify 404

HOST=${HOST:-http://localhost:9090}
RETRIES=${RETRIES:-10}
SLEEP=${SLEEP:-1}

echo "Smoke test target: $HOST"

echo "-> waiting for /healthz (retries=$RETRIES)"
i=0
until curl -fsS "$HOST/healthz" >/dev/null 2>&1; do
  i=$((i+1))
  if [ "$i" -ge "$RETRIES" ]; then
    echo "healthz did not become available after $RETRIES attempts"
    exit 2
  fi
  sleep "$SLEEP"
done
echo "healthz OK"

uniq=$(date +%s)-$RANDOM
email="smoke-$uniq@example.com"
create_payload=$(cat <<EOF
{"username":"smoke-$uniq","email":"$email","password":"password123","role":"admin"}
EOF
)

echo "-> creating user"
create_resp=$(curl -fsS -X POST "$HOST/api/v1/users" -H "Content-Type: application/json" --data-binary "$create_payload" || true)
if [ -z "$create_resp" ]; then
  echo "create request failed (empty response)"
  echo "raw response: $create_resp"
  exit 3
fi

# extract id using python3 if available, otherwise use sed
if command -v python3 >/dev/null 2>&1; then
  id=$(printf "%s" "$create_resp" | python3 -c 'import sys,json
try:
  obj=json.load(sys.stdin)
  print(obj.get("id",""))
except Exception:
  sys.exit(1)') || true
else
  id=$(printf "%s" "$create_resp" | sed -n 's/.*"id"\s*:\s*\([0-9]*\).*/\1/p' || true)
fi

if [ -z "$id" ]; then
  echo "failed to parse id from create response"
  echo "response: $create_resp"
  exit 4
fi
echo "created user id=$id"

echo "-> list users"
list_resp=$(curl -fsS "$HOST/api/v1/users" || true)
echo "$list_resp" | sed -n '1,200p'

echo "-> get user $id"
get_resp=$(curl -fsS "$HOST/api/v1/users/$id" || true)
echo "$get_resp" | sed -n '1,200p'

echo "-> update user $id (username -> smoke-updated-$uniq)"
update_payload=$(cat <<EOF
{"username":"smoke-updated-$uniq"}
EOF
)
update_resp=$(curl -fsS -X PUT "$HOST/api/v1/users/$id" -H "Content-Type: application/json" --data-binary "$update_payload" || true)
if printf "%s" "$update_resp" | grep -q "error"; then
  echo "update returned error: $update_resp"
  exit 5
fi
echo "$update_resp" | sed -n '1,200p'

echo "-> delete user $id"
del_status=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$HOST/api/v1/users/$id" || true)
if [ "$del_status" != "204" ]; then
  echo "expected 204 when deleting, got $del_status"
  exit 6
fi
echo "deleted (204)"

echo "-> get user after delete (expect 404)"
after_status=$(curl -s -o /dev/null -w "%{http_code}" "$HOST/api/v1/users/$id" || true)
if [ "$after_status" != "404" ]; then
  echo "expected 404 after delete, got $after_status"
  exit 7
fi
echo "smoke test passed"

exit 0

### Events CRUD
echo "\n== Events CRUD =="
ev_payload=$(cat <<EOF
{"name":"smoke-event-$uniq","description":"smoke test event"}
EOF
)
echo "-> create event"
ev_resp=$(curl -fsS -X POST "$HOST/api/v1/events" -H "Content-Type: application/json" --data-binary "$ev_payload" || true)
if [ -z "$ev_resp" ]; then echo "event create failed"; exit 8; fi
if command -v python3 >/dev/null 2>&1; then ev_id=$(printf "%s" "$ev_resp" | python3 -c 'import sys,json
try:
  obj=json.load(sys.stdin)
  print(obj.get("id",""))
except Exception:
  sys.exit(1)') || true
else ev_id=$(printf "%s" "$ev_resp" | sed -n 's/.*"id"\s*:\s*\([0-9]*\).*/\1/p' || true); fi
echo "created event id=$ev_id"

echo "-> list events"
curl -fsS "$HOST/api/v1/events" || true

echo "-> get event $ev_id"
curl -fsS "$HOST/api/v1/events/$ev_id" || true

echo "-> update event $ev_id"
ev_upd=$(curl -fsS -X PUT "$HOST/api/v1/events/$ev_id" -H "Content-Type: application/json" --data-binary '{"name":"smoke-event-updated"}' || true)
echo "$ev_upd"

echo "-> delete event $ev_id"
ev_del=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$HOST/api/v1/events/$ev_id" || true)
echo "status: $ev_del"

### Locations CRUD
echo "\n== Locations CRUD =="
loc_payload=$(cat <<EOF
{"user_id":1,"event_id":1,"latitude":-6.2,"longitude":106.8,"timestamp":$(date +%s)}
EOF
)
echo "-> create location"
loc_resp=$(curl -fsS -X POST "$HOST/api/v1/locations" -H "Content-Type: application/json" --data-binary "$loc_payload" || true)
if [ -z "$loc_resp" ]; then echo "location create failed"; exit 9; fi
if command -v python3 >/dev/null 2>&1; then loc_id=$(printf "%s" "$loc_resp" | python3 -c 'import sys,json
try:
  obj=json.load(sys.stdin)
  print(obj.get("id",""))
except Exception:
  sys.exit(1)') || true
else loc_id=$(printf "%s" "$loc_resp" | sed -n 's/.*"id"\s*:\s*\([0-9]*\).*/\1/p' || true); fi
echo "created location id=$loc_id"

echo "-> list locations"
curl -fsS "$HOST/api/v1/locations" || true

echo "-> get location $loc_id"
curl -fsS "$HOST/api/v1/locations/$loc_id" || true

echo "-> update location $loc_id"
loc_upd=$(curl -fsS -X PUT "$HOST/api/v1/locations/$loc_id" -H "Content-Type: application/json" --data-binary '{"latitude":-6.21}' || true)
echo "$loc_upd"

echo "-> delete location $loc_id"
loc_del=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$HOST/api/v1/locations/$loc_id" || true)
echo "status: $loc_del"

echo "all additional resource checks done"
