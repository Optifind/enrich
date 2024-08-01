#!/bin/bash

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL is not set."
  exit 1
fi

# Load database from dump file
psql $DATABASE_URL -f data/db_dumps/latest.sql
