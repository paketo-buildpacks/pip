package pip

import (
	"fmt"
	"os"
	"os/exec"
)

type PIP struct {
}

func (p PIP) InstallVendor(requirementsPath, location, vendorDir string) error {
	cmd := exec.Command(
		"python",
		"-m",
		"pip",
		"install",
		"-r",
		requirementsPath,
		"--ignore-installed",
		"--exists-action=w",
		"--no-index",
		"--find-links=file://"+vendorDir,
		"--compile",
		"--user",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", location))
	return cmd.Run()
}

func (p PIP) Install(requirementsPath, location, cacheDir string) error {
	cmd := exec.Command(
		"python",
		"-m",
		"pip",
		"install",
		"-r",
		requirementsPath,
		"--upgrade",
		"--upgrade-strategy",
		"only-if-needed",
		"--ignore-installed",
		"--exists-action=w",
		"--cache-dir="+cacheDir,
		"--compile",
		"--user",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", location))
	return cmd.Run()
}
