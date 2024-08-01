#!/bin/bash

username=$(echo $DATABASE_URL | sed -n ':\/\/([^:]+):')
password=$(echo $DATABASE_URL | sed -n ':([^:]+)@')
host=$(echo $DATABASE_URL | sed -n '@([^:]+)')
port=$(echo $DATABASE_URL | sed -n ':([0-9]+)\/')
dbname=$(echo $DATABASE_URL | sed -n '\/([^\/]+)$')

# Password is exported to environment variable so pg_restore can use it
export PGPASSWORD=$password

echo "$username"
echo "$password"
echo "$host"
echo "$port"
echo "$dbname"

# Database is restored with dump file
pg_restore --verbose --clean --no-acl --no-owner -h $host -p $port -U $username -d $dbname -f data/db_dumps/latest.dump

# PGPASSWORD is removed from environment variables
unset PGPASSWORD
