package db

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/snowflakedb/gosnowflake"
	_ "github.com/snowflakedb/gosnowflake"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		panic("Missing required env var: " + key)
	}
	return v
}

func loadRSAPrivateKeyFromPEM(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM: could not decode")
	}

	// Try PKCS#8 first
	if keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if k, ok := keyAny.(*rsa.PrivateKey); ok {
			return k, nil
		}
		return nil, fmt.Errorf("private key is not RSA")
	}

	// Fallback to PKCS#1
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}

	return nil, fmt.Errorf("failed to parse RSA private key (PKCS#8 or PKCS#1)")
}

func Connect() (*sql.DB, error) {
	account := mustEnv("SNOWFLAKE_ACCOUNT")
	user := mustEnv("SNOWFLAKE_USER")
	privateKeyPath := mustEnv("SNOWFLAKE_PRIVATE_KEY_PATH")

	role := os.Getenv("SNOWFLAKE_ROLE")
	warehouse := os.Getenv("SNOWFLAKE_WAREHOUSE")
	database := os.Getenv("SNOWFLAKE_DATABASE")
	schema := os.Getenv("SNOWFLAKE_SCHEMA")

	privateKey, err := loadRSAPrivateKeyFromPEM(privateKeyPath)
	if err != nil {
		return nil, err
	}

	cfg := &gosnowflake.Config{
		Account:       account,
		User:          user,
		Authenticator: gosnowflake.AuthTypeJwt, // key-pair auth
		PrivateKey:    privateKey,
		Role:          role,
		Warehouse:     warehouse,
		Database:      database,
		Schema:        schema,
	}

	dsn, err := gosnowflake.DSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build DSN: %w", err)
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
