package pip

import (
	"os"
	"os/exec"
)

type PIP struct {
	// TODO add logger and runner interface
}

func (p PIP) InstallVendor(requirementsPath, location, vendorDir string) error {
	installArgs := []string{"-m", "pip", "install", "-r", requirementsPath, "--ignore-installed", "--no-warn-script-location", "--exists-action=w", "--no-index", "--find-links=file://"+vendorDir, "--target="+location}
	cmd := exec.Command("python", installArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p PIP) Install(requirementsPath, location string) error {
	cmd := exec.Command("python", "-m", "pip", "install", "-r", requirementsPath, "--ignore-installed", "--exists-action=w", "--target="+location)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
