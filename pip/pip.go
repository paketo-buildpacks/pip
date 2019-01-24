package pip

import (
	"os"
	"os/exec"
)

type PIP struct {
	// TODO add logger and runner interface
}

func (p PIP) Install(requirementsPath, location string) error {
	// TODO can --src be replaced by --prefix
	cmd := exec.Command("pip", "install", "-r", requirementsPath, "--ignore-installed", "--exists-action=w", "--src="+location)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
