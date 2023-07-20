#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER muxwaf WITH PASSWORD 'muxwaf@password';
    CREATE DATABASE muxwaf;
    GRANT ALL PRIVILEGES ON DATABASE muxwaf TO muxwaf;
    CREATE TYPE origin_protocol  AS ENUM ('http', 'https', 'follow');
EOSQL