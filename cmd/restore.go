package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/zaniluca/pgs3/internal/s3"

	"github.com/spf13/cobra"
)

type RestoreConfig struct {
	Latest bool
}

var restoreCfg RestoreConfig

const backupFile = "backup.dump"

func init() {
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a backup from S3",
		RunE:  restoreAction,
	}

	setCommonFlags(restoreCmd)

	// Restore flags
	restoreCmd.Flags().BoolVar(&restoreCfg.Latest, "latest", false, "Restore the latest backup (required)")
	restoreCmd.MarkFlagRequired("latest")

	rootCmd.AddCommand(restoreCmd)
}

func restoreAction(cmd *cobra.Command, args []string) error {
	if restoreCfg.Latest {
		return restoreLatestBackup()
	}

	return nil
}

func restoreLatestBackup() error {
	s3Client, err := s3.NewClient(envCfg.AwsAccessKeyId, envCfg.AwsSecretAccessKey, envCfg.AwsRegion, envCfg.AwsS3Endpoint)
	if err != nil {
		return fmt.Errorf("error creating S3 client: %v", err)
	}

	latestBackupKey, err := s3Client.LatestBackup(envCfg.AwsS3Bucket)
	if err != nil {
		return fmt.Errorf("error downloading latest backup: %v", err)
	}

	_, err = s3Client.DownloadFile(envCfg.AwsS3Bucket, latestBackupKey, backupFile)
	defer func() {
		err = os.Remove(backupFile)
		if err != nil {
			err = fmt.Errorf("error removing local dump file: %v", err)
		}
	}()

	if err != nil {
		return fmt.Errorf("error downloading latest backup: %v", err)
	}
	fmt.Printf("Downloaded latest backup from S3: %s\n", latestBackupKey)

	err = restorePgDump(backupFile)
	if err != nil {
		return fmt.Errorf("error restoring backup: %v", err)
	}

	fmt.Println("Latest backup restored")
	return err
}

func restorePgDump(file string) error {
	cmd := exec.Command("pg_restore",
		"-h", envCfg.PostgresHost,
		"-p", envCfg.PostgresPort,
		"-U", envCfg.PostgresUser,
		"-d", envCfg.PostgresDb,
		"--clean",
		"--if-exists",
		file,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", envCfg.PostgresPassword))
	return cmd.Run()
}
