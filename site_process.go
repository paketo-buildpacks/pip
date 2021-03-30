package pip

import (
	"bytes"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/pexec"
)

// SiteProcess implements the Executable interface.
type SiteProcess struct {
	executable Executable
}

// NewSiteProcess creates an instance of the SiteProcess given an Executable that runs `python`
func NewSiteProcess(executable Executable) SiteProcess {
	return SiteProcess{
		executable: executable,
	}
}

// Execute runs a python command to locate the site packages within the pip targetLayerPath.
func (p SiteProcess) Execute(targetLayerPath string) (string, error) {
	buffer := bytes.NewBuffer(nil)
	sitePackagesPath := bytes.NewBuffer(nil)

	err := p.executable.Execute(pexec.Execution{
		// Run the python -m site --user-site to locate the user level site-packages.
		Args: []string{"-m", "site", "--user-site"},
		// Set the PYTHONUSERBASE to ensure that we are looking at the pip layer for user level packages.
		Env:    append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: sitePackagesPath,
		Stderr: buffer,
	})

	if err != nil {
		return "", fmt.Errorf("failed to locate site packages:\n%s\nerror: %w", buffer.String(), err)
	}

	return sitePackagesPath.String(), nil
}
