#!/usr/bin/env bash
set -euo pipefail

GOLINKS="./golinks"
PORT=8080
BASE="http://localhost:${PORT}"
CONFIG='postgres://coder:coder@localhost:5432/postgres?sslmode=disable'
MIGRATE_FILE="./links_test.txt"

# Build
go build -o golinks ./cmd/golinks

# Prepare test file
cat > "$MIGRATE_FILE" <<'EOF'
test https://example.com
foo https://foo.com
EOF

# Start server in background
./golinks \
  --storage POSTGRES \
  --config "$CONFIG" \
  --migrate-from "$MIGRATE_FILE" \
  -port "$PORT" \
  >golinks.log 2>&1 &
SERVER_PID=$!

# Wait for server to be ready
for i in $(seq 1 10); do
  if curl -s "$BASE/api/v1/all" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

echo "========================================="
echo "  HAPPY PATH TESTS"
echo "========================================="

echo ""
echo "--- 1. POST /api/v1/links/testlink ---"
curl -s -X POST "$BASE/api/v1/links/testlink" \
     -H "Content-Type: application/json" \
     -d '{"target":"https://example.org"}'
echo ""
echo "Status: $(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE/api/v1/links/testlink" \
     -H "Content-Type: application/json" \
     -d '{"target":"https://example.org"}')"

echo ""
echo "--- 2. GET /api/v1/links/testlink ---"
curl -s "$BASE/api/v1/links/testlink"
echo ""

echo ""
echo "--- 3. GET /api/v1/all ---"
curl -s "$BASE/api/v1/all" | python3 -m json.tool 2>/dev/null || curl -s "$BASE/api/v1/all"
echo ""

echo ""
echo "--- 4. DELETE /api/v1/links/testlink ---"
curl -s -X DELETE "$BASE/api/v1/links/testlink" -w "\nStatus: %{http_code}\n"

echo ""
echo "--- 5. GET /api/v1/links/testlink (should be 404) ---"
curl -s -w "\nStatus: %{http_code}\n" "$BASE/api/v1/links/testlink"

echo ""
echo "========================================="
echo "  NON-HAPPY PATH TESTS"
echo "========================================="

echo ""
echo "--- 6. POST with invalid JSON ---"
curl -s -o /dev/null -w "Status: %{http_code}\n" -X POST "$BASE/api/v1/links/invalid" \
     -H "Content-Type: application/json" \
     -d 'not a json'

echo ""
echo "--- 7. GET non-existent link ---"
curl -s -w "\nStatus: %{http_code}\n" "$BASE/api/v1/links/doesnotexist"

echo ""
echo "========================================="
echo "  SERVER LOGS"
echo "========================================="
cat golinks.log

# Cleanup
kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
rm -f "$MIGRATE_FILE" golinks.log

echo ""
echo "Done."
