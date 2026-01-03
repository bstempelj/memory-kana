#!/bin/bash

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE TABLE IF NOT EXISTS "player_times" (
		"id" SERIAL PRIMARY KEY,
		"player" VARCHAR(50) NOT NULL,
		"time" TIME NOT NULL
	);
EOSQL
