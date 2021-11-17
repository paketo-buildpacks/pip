package pip_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/pexec"
	pip "github.com/paketo-buildpacks/pip"
	"github.com/paketo-buildpacks/pip/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPipInstallProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		srcLayerPath    string
		targetLayerPath string
		executable      *fakes.Executable

		pipInstallProcess pip.PipInstallProcess
	)

	it.Before(func() {
		var err error
		srcLayerPath, err = os.MkdirTemp("", "pip-source")
		Expect(err).NotTo(HaveOccurred())

		targetLayerPath, err = os.MkdirTemp("", "pip")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}

		pipInstallProcess = pip.NewPipInstallProcess(executable)
	})

	context("Execute", func() {
		context("there is a pip dependency to install", func() {
			it("installs it to the pip layer", func() {
				err := pipInstallProcess.Execute(srcLayerPath, targetLayerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath))))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"-m", "pip", "install", srcLayerPath, "--user", fmt.Sprintf("--find-links=%s", srcLayerPath)}))
			})
		})

		context("failure cases", func() {
			context("the pip install process fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "stdout output")
						fmt.Fprintln(execution.Stderr, "stderr output")
						return errors.New("installing pip failed")
					}
				})

				it("returns an error", func() {
					err := pipInstallProcess.Execute(srcLayerPath, targetLayerPath)
					Expect(err).To(MatchError(ContainSubstring("installing pip failed")))
					Expect(err).To(MatchError(ContainSubstring("stdout output")))
					Expect(err).To(MatchError(ContainSubstring("stderr output")))
				})
			})
		})
	})
}
