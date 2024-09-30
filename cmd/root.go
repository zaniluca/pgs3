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

var envCfg EnvConfig

var (
	rootCmd = &cobra.Command{
		Use:   "pgs3",
		Short: "Backup and restore PostgreSQL databases to/from S3",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}
			return nil
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			envCfg.PostgresDb = getEnvOrFlag(cmd, "POSTGRES_DB", "postgres-db")
			envCfg.PostgresUser = getEnvOrFlag(cmd, "POSTGRES_USER", "postgres-user")
			envCfg.PostgresPassword = getEnvOrFlag(cmd, "POSTGRES_PASSWORD", "postgres-password")
			envCfg.PostgresHost = getEnvOrFlag(cmd, "POSTGRES_HOST", "postgres-host")
			envCfg.PostgresPort = getEnvOrFlag(cmd, "POSTGRES_PORT", "postgres-port")

			envCfg.AwsS3Endpoint = getEnvOrFlag(cmd, "AWS_S3_ENDPOINT", "s3-endpoint")
			envCfg.AwsS3Bucket = getEnvOrFlag(cmd, "AWS_S3_BUCKET", "s3-bucket")
			envCfg.AwsRegion = getEnvOrFlag(cmd, "AWS_REGION", "aws-region")
			envCfg.AwsAccessKeyId = getEnvOrFlag(cmd, "AWS_ACCESS_KEY_ID", "aws-access-key-id")
			envCfg.AwsSecretAccessKey = getEnvOrFlag(cmd, "AWS_SECRET_ACCESS_KEY", "aws-secret-access-key")
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

// getEnvOrFlag checks for an environment variable first, then falls back to the flag value
func getEnvOrFlag(cmd *cobra.Command, envName, flagName string) string {
	envValue := os.Getenv("PGS3_" + envName)
	cmd.Flags().Set(flagName, envValue)
	if envValue != "" {
		return envValue
	}

	return cmd.Flag(flagName).Value.String()
}

// setCommonFlags sets the flags common to every subcommand (aws and postgres related credentials and configuration options)
func setCommonFlags(cmd *cobra.Command) {
	// Postgres flags
	cmd.Flags().StringVarP(&envCfg.PostgresDb, "postgres-db", "d", "", "PostgreSQL Database name, can be set with PGS3_POSTGRES_DB env (required)")
	cmd.Flags().StringVarP(&envCfg.PostgresUser, "postgres-user", "U", "", "PostgreSQL user, can be set with PGS3_POSTGRES_USER env (required)")
	cmd.Flags().StringVarP(&envCfg.PostgresPassword, "postgres-password", "P", "", "PostgreSQL password, can be set with PGS3_POSTGRES_PASSWORD env (required)")
	cmd.Flags().StringVarP(&envCfg.PostgresHost, "postgres-host", "H", "localhost", "PostgreSQL host, can be set with PGS3_POSTGRES_HOST env")
	cmd.Flags().StringVarP(&envCfg.PostgresPort, "postgres-port", "p", "5432", "PostgreSQL port, can be set with PGS3_POSTGRES_PORT env")
	cmd.MarkFlagRequired("postgres-db")
	cmd.MarkFlagRequired("postgres-user")
	cmd.MarkFlagRequired("postgres-password")
	// AWS flags
	cmd.Flags().StringVar(&envCfg.AwsS3Endpoint, "s3-endpoint", "", "AWS S3 custom endpoint, can be set with PGS3_AWS_S3_ENDPOINT env")
	cmd.Flags().StringVarP(&envCfg.AwsS3Bucket, "s3-bucket", "b", "", "AWS S3 Bucket name, can be set with PGS3_AWS_S3_BUCKET env (required)")
	cmd.Flags().StringVarP(&envCfg.AwsRegion, "aws-region", "r", "", "AWS Region the bucket is stored, can be set with PGS3_AWS_REGION env (required)")
	cmd.Flags().StringVar(&envCfg.AwsAccessKeyId, "aws-access-key-id", "", "AWS Key ID, can be set with PGS3_AWS_ACCESS_KEY_ID env (required)")
	cmd.Flags().StringVar(&envCfg.AwsSecretAccessKey, "aws-secret-access-key", "", "AWS Secret Key, can be set with PGS3_AWS_SECRET_ACCESS_KEY env (required)")
	cmd.MarkFlagRequired("s3-bucket")
	cmd.MarkFlagRequired("aws-region")
	cmd.MarkFlagRequired("aws-access-key-id")
	cmd.MarkFlagRequired("aws-secret-access-key")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Print error message in bold red
		fmt.Printf("\033[31;1m%v\033[0m\n", err)
		os.Exit(1)
	}
}
