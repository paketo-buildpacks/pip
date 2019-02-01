package pip

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/libcfbuildpack/helper"
)

type Logger interface {
	Info(format string, args ...interface{})
}

type PIP struct {
	Logger Logger
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
	if cacheExists, err := helper.FileExists(cacheDir); err != nil {
		return err
	} else if cacheExists {
		p.Logger.Info("Reusing existing pip cache")
	}

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
