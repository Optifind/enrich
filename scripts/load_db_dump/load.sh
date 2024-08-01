#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL is not set."
  exit 1
fi

# Parse the DATABASE_URL to extract components using the provided regex patterns
username=$(echo $DATABASE_URL | sed -n 's/^postgres:\/\/\([^:]\+\):.*$/\1/p')
password=$(echo $DATABASE_URL | sed -n 's/^postgres:\/\/[^:]\+:\([^@]\+\)@.*$/\1/p')
host=$(echo $DATABASE_URL | sed -n 's/^postgres:\/\/[^:]\+:[^@]\+@\([^:]\+\).*$/\1/p')
port=$(echo $DATABASE_URL | sed -n 's/^postgres:\/\/[^:]\+:[^@]\+@[^:]\+:\([0-9]\+\).*$/\1/p')
dbname=$(echo $DATABASE_URL | sed -n 's/^postgres:\/\/[^:]\+:[^@]\+@[^:]\+:[0-9]\+\/\([^\/]\+\).*$/\1/p')

# Export password to environment variable so pg_restore can use it
export PGPASSWORD=$password

# Drop and recreate the database to ensure a clean slate
# Use caution with these commands as they will erase the existing database
psql -h $host -p $port -U $username -d postgres -c "DROP DATABASE IF EXISTS $dbname;"
psql -h $host -p $port -U $username -d postgres -c "CREATE DATABASE $dbname;"


# Run pg_restore with the parsed components
pg_restore --verbose --clean --no-acl --no-owner -h $host -p $port -U $username -d $dbname data/db_dumps/latest.dump

# Unset PGPASSWORD
unset PGPASSWORD
