package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zaniluca/pgs3/internal/s3"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
)

type BackupConfig struct {
	Schedule         string
	RestoreOnStartup bool
	KeepDays         int
	PgDumpExtraOpts  string
}

var backupConfig BackupConfig

func init() {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup and upload to S3",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if backupConfig.RestoreOnStartup && backupConfig.Schedule == "" {
				return fmt.Errorf("the --restore-on-startup flag can only be used if --schedule is set")
			}
			return nil
		},
		RunE: backupAction,
	}

	// Postgres flags
	backupCmd.Flags().StringVarP(&envConfig.PostgresDb, "postgres-db", "d", "", "PostgreSQL Database name, or set POSTGRES_DB env  (required)")
	backupCmd.Flags().StringVarP(&envConfig.PostgresUser, "postgres-user", "U", "", "PostgreSQL user, or set POSTGRES_USER env (required)")
	backupCmd.Flags().StringVarP(&envConfig.PostgresPassword, "postgres-password", "P", "", "PostgreSQL password, or set POSTGRES_PASSWORD env  (required)")
	backupCmd.Flags().StringVarP(&envConfig.PostgresHost, "postgres-host", "H", "localhost", "PostgreSQL host")
	backupCmd.Flags().StringVarP(&envConfig.PostgresPort, "postgres-port", "p", "5432", "PostgreSQL port")
	backupCmd.MarkFlagRequired("postgres-db")
	backupCmd.MarkFlagRequired("postgres-user")
	backupCmd.MarkFlagRequired("postgres-password")
	// AWS flags
	backupCmd.Flags().StringVar(&envConfig.AwsS3Endpoint, "s3-endpoint", "", "AWS S3 custom endpoint")
	backupCmd.Flags().StringVarP(&envConfig.AwsS3Bucket, "s3-bucket", "b", "", "AWS S3 Bucket name (required)")
	backupCmd.Flags().StringVarP(&envConfig.AwsRegion, "aws-region", "r", "", "AWS Region the bucket is stored (required)")
	backupCmd.Flags().StringVar(&envConfig.AwsAccessKeyId, "aws-access-key-id", "", "AWS Key ID, or set AWS_ACCESS_KEY_ID env  (required)")
	backupCmd.Flags().StringVar(&envConfig.AwsSecretAccessKey, "aws-secret-access-key", "", "AWS Secret Key, or set AWS_SECRET_ACCESS_KEY env  (required)")
	backupCmd.MarkFlagRequired("s3-bucket")
	backupCmd.MarkFlagRequired("aws-region")
	backupCmd.MarkFlagRequired("aws-access-key-id")
	backupCmd.MarkFlagRequired("aws-secret-access-key")
	// Backup flags
	backupCmd.Flags().StringVar(&backupConfig.Schedule, "schedule", "", "Cron schedule for periodic backups in crontab format (e.g. '0 0 * * *')")
	backupCmd.Flags().BoolVar(&backupConfig.RestoreOnStartup, "restore-on-startup", false, "Before starting the backup, restore the database from the latest backup in S3 (requires --schedule)")
	backupCmd.Flags().IntVar(&backupConfig.KeepDays, "keep-days", 0, "Number of days to keep backups in S3")
	backupCmd.Flags().StringVar(&backupConfig.PgDumpExtraOpts, "pgdump-extra-opts", "", "Extra options to pass to pg_dump")

	rootCmd.AddCommand(backupCmd)
}

func backupAction(cmd *cobra.Command, args []string) error {
	if backupConfig.Schedule != "" {
		if backupConfig.RestoreOnStartup {
			restoreLatestBackup()
		}

		s, err := gocron.NewScheduler()
		if err != nil {
			return fmt.Errorf("error creating scheduler: %v", err)
		}

		_, err = s.NewJob(gocron.CronJob(backupConfig.Schedule, false), gocron.NewTask(performBackup))
		if err != nil {
			if errors.Is(err, gocron.ErrCronJobParse) {
				return fmt.Errorf("schedule is invalid: %v", backupConfig.Schedule)
			}
			return fmt.Errorf("error setting up cron job: %v", err)
		}

		s.Start()
		fmt.Printf("Starting periodic backup with schedule: %s\n", backupConfig.Schedule)

		select {} // Block forever
	} else {
		return performBackup()
	}
}

func performBackup() error {
	fmt.Println("Creating backup...")
	dumpFile, err := createPgDump(envConfig.PostgresHost, envConfig.PostgresPort, envConfig.PostgresDb, envConfig.PostgresUser, envConfig.PostgresPassword, "")
	if err != nil {
		return fmt.Errorf("error creating PostgreSQL dump: %v", err)
	}

	s3Client, err := s3.NewClient(envConfig.AwsAccessKeyId, envConfig.AwsSecretAccessKey, envConfig.AwsRegion, envConfig.AwsS3Endpoint)
	if err != nil {
		return fmt.Errorf("error creating S3 client: %v", err)
	}

	err = s3Client.UploadFile(envConfig.AwsS3Bucket, dumpFile)
	if err != nil {
		return fmt.Errorf("error uploading to S3: %v", err)
	}

	// Remove local dump file
	err = os.Remove(dumpFile)
	if err != nil {
		return fmt.Errorf("error removing local dump file: %v", err)
	}

	if backupConfig.KeepDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -backupConfig.KeepDays)
		err := s3Client.RemoveOldBackups(envConfig.AwsS3Bucket, cutoffDate)
		if err != nil {
			return fmt.Errorf("error removing old backups: %v", err)
		}
	}

	fmt.Println("Backup complete.")
	return nil
}

func createPgDump(host, port, dbname, user, password, extraOpts string) (string, error) {
	dumpFile := fmt.Sprintf("%s_%s.dump", dbname, time.Now().Format("2006-01-02T15:04:05"))
	cmd := exec.Command("pg_dump",
		"--format=custom",
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbname,
		"-f", dumpFile,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	if extraOpts != "" {
		cmd.Args = append(cmd.Args, strings.Split(extraOpts, " ")...)
	}
	return dumpFile, cmd.Run()
}
