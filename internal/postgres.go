package pgs3

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func CreatePgDump(host, port, dbname, user, password, extraOpts string) (string, error) {
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

func RestorePgDump(host, port, dbname, user, password, file string) error {
	cmd := exec.Command("pg_restore",
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbname,
		"--clean",
		"--if-exists",
		file,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	return cmd.Run()
}
