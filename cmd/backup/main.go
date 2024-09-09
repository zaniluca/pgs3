package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"pg-s3-backup/internal/env"
	"pg-s3-backup/internal/s3"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
)

var (
	schedule         string
	restoreOnStartup bool
	rootCmd          = &cobra.Command{
		Use:   "pg-s3-backup",
		Short: "Backup and restore PostgreSQL databases to/from S3",
	}
)

func init() {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup and upload to S3",
		Run:   backupAction,
	}
	backupCmd.Flags().StringVar(&schedule, "schedule", "", "Cron schedule for periodic backups")
	backupCmd.Flags().BoolVar(&restoreOnStartup, "restore-on-startup", false, "Before starting the backup, restore the database from the latest backup in S3")
	rootCmd.AddCommand(backupCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func backupAction(cmd *cobra.Command, args []string) {
	if schedule != "" {
		if restoreOnStartup {
			// Restore the database before starting the backup
		}
		s, err := gocron.NewScheduler()
		if err != nil {
			log.Fatalf("Error creating scheduler: %v", err)
		}
		_, err = s.NewJob(gocron.CronJob(schedule, true), gocron.NewTask(performBackup))
		if err != nil {
			log.Fatalf("Error setting up cron job: %v", err)
		}
		fmt.Printf("Starting periodic backup with schedule: %s\n", schedule)

		select {} // Block forever
	} else {
		performBackup()
	}
}
func performBackup() {
	fmt.Println("Creating backup...")

	env, err := env.Load()
	if err != nil {
		log.Fatalf("Error loading environment: %v", err)
	}

	dumpFile, err := createPgDump(env)
	if err != nil {
		log.Fatalf("Error creating PostgreSQL dump: %v", err)
	}

	s3Client, err := s3.NewClient(env.S3AccessKeyID, env.S3SecretAccessKey, env.S3Region, env.S3Endpoint)
	if err != nil {
		log.Fatalf("Error creating S3 client: %v", err)
	}

	err = s3Client.UploadFile(env.S3Bucket, dumpFile)
	if err != nil {
		log.Fatalf("Error uploading to S3: %v", err)
	}

	// Remove local dump file
	os.Remove(dumpFile)

	if env.BackupKeepDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -env.BackupKeepDays)
		err := s3Client.RemoveOldBackups(env.S3Bucket, cutoffDate)
		if err != nil {
			log.Printf("Error removing old backups: %v", err)
		}
	}

	fmt.Println("Backup complete.")
}

func createPgDump(env *env.Env) (string, error) {
	dumpFile := fmt.Sprintf("%s_%s.dump", env.PostgresDatabase, time.Now().Format("2006-01-02T15:04:05"))
	cmd := exec.Command("pg_dump",
		"--format=custom",
		"-h", env.PostgresHost,
		"-p", env.PostgresPort,
		"-U", env.PostgresUser,
		"-d", env.PostgresDatabase,
		"-f", dumpFile,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", env.PostgresPassword))
	if env.PgDumpExtraOpts != "" {
		cmd.Args = append(cmd.Args, strings.Split(env.PgDumpExtraOpts, " ")...)
	}
	return dumpFile, cmd.Run()
}
