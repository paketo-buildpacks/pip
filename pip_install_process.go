package pip

import (
	"bytes"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go

// Executable defines the interface for invoking an executable.
type Executable interface {
	Execute(pexec.Execution) error
}

// PipInstallProcess implements the InstallProcess interface.
type PipInstallProcess struct {
	executable Executable
}

// NewPipInstallProcess creates an instance of the PipInstallProcess given an Executable that runs `python`.
func NewPipInstallProcess(executable Executable) PipInstallProcess {
	return PipInstallProcess{
		executable: executable,
	}
}

// Execute installs the pip binary from source code located in the given srcPath into the a layer path designated by targetLayerPath.
func (p PipInstallProcess) Execute(srcPath, targetLayerPath string) error {
	buffer := bytes.NewBuffer(nil)

	err := p.executable.Execute(pexec.Execution{
		Args:   []string{"-m", "pip", "install", srcPath, "--user"},
		Env:    append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return fmt.Errorf("failed to configure pip:\n%s\nerror: %w", buffer.String(), err)
	}
	return nil
}
