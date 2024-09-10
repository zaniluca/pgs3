package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"pg-s3-backup/internal/env"

	"github.com/spf13/cobra"
)

var (
	timestamp string
	latest    bool
)

func init() {
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a backup from S3",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	restoreCmd.Flags().StringVarP(&timestamp, "timestamp", "t", "", "Timestamp of the backup to restore")
	restoreCmd.Flags().BoolVar(&latest, "latest", false, "Restore the latest backup")

	restoreCmd.MarkFlagsMutuallyExclusive("timestamp", "latest")
	restoreCmd.MarkFlagsOneRequired("timestamp", "latest")

	rootCmd.AddCommand(restoreCmd)
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
