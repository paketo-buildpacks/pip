package pip

import (
	"os"
	"os/exec"
)

type PIP struct {
	// TODO add logger and runner interface
}

func (p PIP) Install(requirementsPath, location string) error {
	cmd := exec.Command("python", "-m", "pip", "-v", "install", "-v", "-r", requirementsPath, "--ignore-installed", "--exists-action=w", "--target="+location)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
