#!/bin/bash

USERNAME=$1
PASSWORD=$2

if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ]; then
    exit 1
fi

DB_PATH="${DB_PATH:-/root/edge_server/server.db}"

QUERY="SELECT password, enabled FROM users WHERE username='$USERNAME'"
RESULT=$(sqlite3 "$DB_PATH" "$QUERY" 2>/dev/null)

if [ -z "$RESULT" ]; then
    exit 1
fi

STORED_HASH=$(echo "$RESULT" | cut -d'|' -f1)
ENABLED=$(echo "$RESULT" | cut -d'|' -f2)

if [ "$ENABLED" != "1" ]; then
    exit 1
fi

cat > /tmp/check_password_$$.py << 'PYEOF'
import sys
import bcrypt

stored_hash = sys.argv[1].encode('utf-8')
password = sys.argv[2].encode('utf-8')

try:
    if bcrypt.checkpw(password, stored_hash):
        sys.exit(0)
    else:
        sys.exit(1)
except:
    sys.exit(1)
PYEOF

if command -v python3 &> /dev/null; then
    python3 -c "import bcrypt" 2>/dev/null
    if [ $? -eq 0 ]; then
        python3 /tmp/check_password_$$.py "$STORED_HASH" "$PASSWORD"
        RET=$?
        rm -f /tmp/check_password_$$.py
        exit $RET
    fi
fi

rm -f /tmp/check_password_$$.py

echo "$PASSWORD" | openssl passwd -apr1 -stdin -salt "$(echo $STORED_HASH | cut -d'$' -f3)" | grep -q "^$STORED_HASH$"
if [ $? -eq 0 ]; then
    exit 0
fi

exit 1
