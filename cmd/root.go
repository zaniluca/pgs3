package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type EnvConfig struct {
	PostgresDb         string
	PostgresUser       string
	PostgresPassword   string
	PostgresHost       string
	PostgresPort       string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsS3Bucket        string
	AwsRegion          string
	AwsS3Endpoint      string
}

var envConfig EnvConfig

var (
	rootCmd = &cobra.Command{
		Use:   "pgs3",
		Short: "Backup and restore PostgreSQL databases to/from S3",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}

			envConfig.PostgresDb = getEnvOrFlag(cmd, "POSTGRES_DB", "postgres-db")
			envConfig.PostgresUser = getEnvOrFlag(cmd, "POSTGRES_USER", "postgres-user")
			envConfig.PostgresPassword = getEnvOrFlag(cmd, "POSTGRES_PASSWORD", "postgres-password")
			envConfig.PostgresHost = getEnvOrFlag(cmd, "POSTGRES_HOST", "postgres-host")
			envConfig.PostgresPort = getEnvOrFlag(cmd, "POSTGRES_PORT", "postgres-port")
			envConfig.AwsS3Endpoint = getEnvOrFlag(cmd, "AWS_S3_ENDPOINT", "s3-endpoint")
			envConfig.AwsS3Bucket = getEnvOrFlag(cmd, "AWS_S3_BUCKET", "s3-bucket")
			envConfig.AwsRegion = getEnvOrFlag(cmd, "AWS_REGION", "aws-region")
			envConfig.AwsAccessKeyId = getEnvOrFlag(cmd, "AWS_ACCESS_KEY_ID", "aws-access-key-id")
			envConfig.AwsSecretAccessKey = getEnvOrFlag(cmd, "AWS_SECRET_ACCESS_KEY", "aws-secret-access-key")

			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

// getEnvOrFlag checks for an environment variable first, then falls back to the flag value
func getEnvOrFlag(cmd *cobra.Command, envName, flagName string) string {
	envValue := os.Getenv("PG_S3_" + envName)
	cmd.Flags().Set(flagName, envValue)
	if envValue != "" {
		return envValue
	}

	return cmd.Flag(flagName).Value.String()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Print error message in bold red
		fmt.Printf("\033[31;1m%v\033[0m\n", err)
		os.Exit(1)
	}
}
