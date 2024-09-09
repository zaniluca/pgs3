package env

import (
	"fmt"
	"os"
	"strconv"
)

type Env struct {
	S3Bucket          string
	PostgresDatabase  string
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Region          string
	PgDumpExtraOpts   string
	BackupKeepDays    int
}

func Load() (*Env, error) {
	env := &Env{
		S3Bucket:          os.Getenv("S3_BUCKET"),
		PostgresDatabase:  os.Getenv("POSTGRES_DATABASE"),
		PostgresHost:      os.Getenv("POSTGRES_HOST"),
		PostgresPort:      os.Getenv("POSTGRES_PORT"),
		PostgresUser:      os.Getenv("POSTGRES_USER"),
		PostgresPassword:  os.Getenv("POSTGRES_PASSWORD"),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Region:          os.Getenv("S3_REGION"),
		PgDumpExtraOpts:   os.Getenv("PGDUMP_EXTRA_OPTS"),
	}

	if env.S3Bucket == "" || env.PostgresDatabase == "" || env.PostgresHost == "" ||
		env.PostgresUser == "" || env.PostgresPassword == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	if env.PostgresPort == "" {
		fmt.Println("POSTGRES_PORT not set, defaulting to 5432")
		env.PostgresPort = "5432"
	}

	if backupKeepDays := os.Getenv("BACKUP_KEEP_DAYS"); backupKeepDays != "" {
		env.BackupKeepDays, _ = strconv.Atoi(backupKeepDays)
	}

	return env, nil
}
