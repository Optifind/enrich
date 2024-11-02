#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "Loading .env file..."

source "./data/secrets.env"

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL is not set."
  exit 1
fi

# Parse the DATABASE_URL to extract components using the provided regex patterns
host=$(echo "$DATABASE_URL" | sed -n 's/.*host=\([^ ]*\).*/\1/p')
port=$(echo "$DATABASE_URL" | sed -n 's/.*port=\([^ ]*\).*/\1/p')
username=$(echo "$DATABASE_URL" | sed -n 's/.*user=\([^ ]*\).*/\1/p')
password=$(echo "$DATABASE_URL" | sed -n 's/.*password=\([^ ]*\).*/\1/p')
dbname=$(echo "$DATABASE_URL" | sed -n 's/.*dbname=\([^ ]*\).*/\1/p')

echo "Hostname: $host"
echo "Port: $port"
echo "Username: $username"
echo "Password: $password"
echo "Database: $dbname"

# Export password to environment variable so pg_restore can use it
export PGPASSWORD=$password

dump_filepath="data/db_dumps/latest.dump"

# Pre-drop the extension with cascade to avoid dependency issues
echo "Dropping extensions to avoid dependency issues..."
psql "postgresql://$username:$password@$host:$port/$dbname?sslmode=disable" <<EOF
DO
\$do\$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        EXECUTE 'DROP EXTENSION vector CASCADE';
    END IF;
END
\$do\$;
EOF

# Run pg_restore with the parsed components and handle extension comments
echo "Restoring database from dump file..."
pg_restore --verbose --clean --if-exists --no-acl --no-owner -h "$host" -p "$port" -U "$username" -d "$dbname" "$dump_filepath" || true

# Reapply extension comments (ignore errors if not a superuser)
psql "postgresql://$username:$password@$host:$port/$dbname?sslmode=disable" <<EOF
DO
\$do\$
BEGIN
    BEGIN
        EXECUTE 'COMMENT ON EXTENSION vector IS ''vector data type and ivfflat and hnsw access methods''';
    EXCEPTION
        WHEN insufficient_privilege THEN
            RAISE NOTICE 'Skipping COMMENT ON EXTENSION due to insufficient privileges';
    END;
END
\$do\$;
EOF

# Unset PGPASSWORD
unset PGPASSWORD
