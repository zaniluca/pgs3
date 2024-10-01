package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
	pgs3 "github.com/zaniluca/pgs3/internal"
)

type BackupConfig struct {
	Schedule         string
	RestoreOnStartup bool
	KeepDays         int
	PgDumpExtraOpts  string
}

var backupCfg BackupConfig

func init() {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup and upload to S3",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if backupCfg.RestoreOnStartup && backupCfg.Schedule == "" {
				return fmt.Errorf("the --restore-on-startup flag can only be used if --schedule is set")
			}
			return nil
		},
		RunE: backupAction,
	}

	setCommonFlags(backupCmd)

	// Backup flags
	backupCmd.Flags().StringVar(&backupCfg.Schedule, "schedule", "", "Cron schedule for periodic backups in crontab format (e.g. '0 0 * * *')")
	backupCmd.Flags().BoolVar(&backupCfg.RestoreOnStartup, "restore-on-startup", false, "Before starting the backup, restore the database from the latest backup in S3 (requires --schedule)")
	backupCmd.Flags().IntVar(&backupCfg.KeepDays, "keep-days", 0, "Number of days to keep backups in S3")
	backupCmd.Flags().StringVar(&backupCfg.PgDumpExtraOpts, "pgdump-extra-opts", "", "Extra options to pass to pg_dump")

	rootCmd.AddCommand(backupCmd)
}

func backupAction(cmd *cobra.Command, args []string) error {
	if backupCfg.Schedule != "" {
		if backupCfg.RestoreOnStartup {
			restoreLatestBackup()
		}

		s, err := gocron.NewScheduler()
		if err != nil {
			return fmt.Errorf("error creating scheduler: %v", err)
		}

		_, err = s.NewJob(gocron.CronJob(backupCfg.Schedule, false), gocron.NewTask(performBackup))
		if err != nil {
			if errors.Is(err, gocron.ErrCronJobParse) {
				return fmt.Errorf("schedule is invalid: %v", backupCfg.Schedule)
			}
			return fmt.Errorf("errcan be set withting up cron job: %v", err)
		}

		s.Start()
		fmt.Printf("Starting periodic backup with schedule: %s\n", backupCfg.Schedule)

		select {} // Block forever
	} else {
		return performBackup()
	}
}

func performBackup() error {
	fmt.Println("Creating backup...")
	dumpFile, err := pgs3.CreatePgDump(envCfg.PostgresHost, envCfg.PostgresPort, envCfg.PostgresDb, envCfg.PostgresUser, envCfg.PostgresPassword, "")
	defer func() {
		err = os.Remove(dumpFile)
		if err != nil {
			err = fmt.Errorf("error removing local dump file: %v", err)
		}
	}()

	if err != nil {
		return fmt.Errorf("error creating PostgreSQL dump: %v", err)
	}

	s3Client, err := pgs3.NewS3Client(envCfg.AwsAccessKeyId, envCfg.AwsSecretAccessKey, envCfg.AwsRegion, envCfg.AwsS3Endpoint)
	if err != nil {
		return fmt.Errorf("error creating S3 client: %v", err)
	}

	err = s3Client.UploadFile(envCfg.AwsS3Bucket, dumpFile)
	if err != nil {
		return fmt.Errorf("error uploading to S3: %v", err)
	}

	if backupCfg.KeepDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -backupCfg.KeepDays)
		fmt.Printf("Removing backups older than %s\n", cutoffDate)
		err := s3Client.RemoveOldBackups(envCfg.AwsS3Bucket, cutoffDate)
		if err != nil {
			return fmt.Errorf("error removing old backups: %v", err)
		}
	}

	fmt.Println("Backup complete.")
	return err
}
