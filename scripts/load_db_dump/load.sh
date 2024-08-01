#!/bin/bash

# Extract the database URL components
# DATABASE_URL format: postgres://username:password@hostname:port/dbname
url=$DATABASE_URL

# Use `awk` or `psql` to parse the URL
username=$(echo $url | awk -F[@:/] '{print $4}')
password=$(echo $url | awk -F[@:/] '{print $5}')
hostname=$(echo $url | awk -F[@:/] '{print $6}')
port=$(echo $url | awk -F[@:/] '{print $7}')
dbname=$(echo $url | awk -F[@:/] '{print $8}')

# Run pg_restore with the parsed components
PGPASSWORD=$password pg_restore --verbose --clean --no-acl --no-owner -h $hostname -U $username -d $dbname data/db_dumps/latest.dump
