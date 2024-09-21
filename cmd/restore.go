package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/zaniluca/pg-s3-toolkit/internal/env"
	"github.com/zaniluca/pg-s3-toolkit/internal/s3"

	"github.com/spf13/cobra"
)

var (
	latest bool
)

const backupFile = "backup.dump"

func init() {
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a backup from S3",
		Run:   restoreAction,
	}
	restoreCmd.Flags().BoolVar(&latest, "latest", false, "Restore the latest backup (required)")

	restoreCmd.MarkFlagRequired("latest")
	rootCmd.AddCommand(restoreCmd)
}

func restoreAction(cmd *cobra.Command, args []string) {
	if latest {
		restoreLatestBackup()

		return
	}
}

func restoreLatestBackup() {
	env, err := env.Load()
	if err != nil {
		log.Printf("Error loading environment: %v\n", err)
	}

	s3Client, err := s3.NewClient(env.S3AccessKeyID, env.S3SecretAccessKey, env.S3Region, env.S3Endpoint)
	if err != nil {
		log.Fatalf("Error creating S3 client: %v", err)
	}

	latestBackupKey, err := s3Client.LatestBackup(env.S3Bucket)
	if err != nil {
		log.Printf("Error downloading latest backup: %v\n", err)
	}

	_, err = s3Client.DownloadFile(env.S3Bucket, latestBackupKey, backupFile)
	if err != nil {
		log.Printf("Error downloading latest backup: %v\n", err)
	}
	log.Printf("Downloaded latest backup from S3: %s\n", latestBackupKey)

	err = restorePgDump(backupFile, env)
	if err != nil {
		log.Printf("Error restoring backup: %v\n", err)
	}
	os.Remove(backupFile)

	log.Println("Latest backup restored")
}

func restorePgDump(file string, env *env.Env) error {
	cmd := exec.Command("pg_restore",
		"-h", env.PostgresHost,
		"-p", env.PostgresPort,
		"-U", env.PostgresUser,
		"-d", env.PostgresDatabase,
		"--clean",
		"--if-exists",
		file,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", env.PostgresPassword))
	return cmd.Run()
}
