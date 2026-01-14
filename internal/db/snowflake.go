package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/snowflakedb/gosnowflake"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("Missing required env var: " + key)
	}
	return v
}

func Connect() (*sql.DB, error) {
	account := mustEnv("SNOWFLAKE_ACCOUNT")
	user := mustEnv("SNOWFLAKE_USER")
	password := mustEnv("SNOWFLAKE_PASSWORD")
	warehouse := mustEnv("SNOWFLAKE_WAREHOUSE")
	database := mustEnv("SNOWFLAKE_DATABASE")
	schema := mustEnv("SNOWFLAKE_SCHEMA")

	role := os.Getenv("SNOWFLAKE_ROLE") // optional

	dsn := fmt.Sprintf("%s:%s@%s/%s/%s?warehouse=%s",
		user, password, account, database, schema, warehouse,
	)

	if role != "" {
		dsn += "&role=" + role
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}

	// Live connectivity check
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
