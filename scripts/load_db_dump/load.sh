#!/bin/bash

# Run pg_restore with the parsed components
psql -f data/db_dumps/latest.sql
