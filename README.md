# snowflake-docker-api

Go API that queries Snowflake (3 endpoints) and is ready to run in Docker.

## Endpoints
- GET `/api/sf/time` -> Snowflake current timestamp
- GET `/api/sf/context` -> current account/user/role/warehouse/db/schema
- GET `/api/sf/menu-items?limit=10` -> lists tables from INFORMATION_SCHEMA

## Local run
Create `.env` (not committed) with:
- `PORT=8080`
- `SNOWFLAKE_ACCOUNT`
- `SNOWFLAKE_USER`
- `SNOWFLAKE_ROLE`
- `SNOWFLAKE_WAREHOUSE`
- `SNOWFLAKE_DATABASE`
- `SNOWFLAKE_SCHEMA`
- `SNOWFLAKE_PRIVATE_KEY_PATH=keys/snowflake_key.pem`

## Then run on terminal:
set -a
source .env
set +a
go run ./cmd/api

## Docker run (after installing Docker):
docker compose up --build