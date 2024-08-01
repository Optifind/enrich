# How to install

## Debian

Add PostgreSQL Apt repository

```bash
sudo apt install -y postgresql-common
sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
```

Update Apt repositories

```bash
sudo apt update
```

Install postgresql

```bash
sudo apt install postgresql-<version> postgresql-contrib
```

Create database cluster

```bash
sudo pg_createcluster 12 main --start
```

Install [pgvector](https://github.com/pgvector/pgvector)

```bash
sudo apt update
sudo apt install postgresql-<version>-pgvector
```

```postgresql
CREATE EXTENSION vector;
```

# How to use

## Debian

Access database shell

```bash
sudo -u postgres psql
```

List databases

```postgresql
\l
```

Create new database

```postgresql
CREATE DATABASE <database_name>;
```

Connect to database

```postgresql
\c <database_name>
```

Create new user

```postgresql
CREATE USER <username> WITH ENCRYPTED PASSWORD <password>;
```

Grant privileges

- All privileges on database
- Can create and drop tables

```postgresql
GRANT ALL PRIVILEGES ON <database_name> TO <username>;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO <username>;
```

List tables

```postgresql
\dt
```

Create table

```postgresql
CREATE TABLE <table_name> (<column_name> <type> PRIMARY KEY, <column_name> <type>);
```

Describe table
```postgresql
SELECT 
  column_name, 
  data_type, 
  character_maximum_length, 
  is_nullable
FROM 
  information_schema.columns 
WHERE 
  table_name = '<table_name>';
```
