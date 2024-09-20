package env

import (
	"fmt"
	"log"
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

	if env.S3AccessKeyID == "" || env.S3SecretAccessKey == "" {
		return nil, fmt.Errorf("missing S3 credentials, set S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY environment variables")
	}

	if env.S3Region == "" || env.S3Bucket == "" {
		return nil, fmt.Errorf("missing S3 configuration, set S3_REGION and S3_BUCKET environment variables")
	}

	if env.PostgresDatabase == "" || env.PostgresUser == "" || env.PostgresPassword == "" {
		return nil, fmt.Errorf("missing PostgreSQL configuration, set POSTGRES_DATABASE, POSTGRES_USER and POSTGRES_PASSWORD environment variables")
	}

	if env.PostgresHost == "" {
		log.Println("POSTGRES_HOST not set, defaulting to localhost")
		env.PostgresHost = "localhost"
	}

	if env.PostgresPort == "" {
		log.Println("POSTGRES_PORT not set, defaulting to 5432")
		env.PostgresPort = "5432"
	}

	if backupKeepDays := os.Getenv("BACKUP_KEEP_DAYS"); backupKeepDays != "" {
		env.BackupKeepDays, _ = strconv.Atoi(backupKeepDays)
	}

	return env, nil
}
