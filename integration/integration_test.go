package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	bpDir, pythonURI, pipURI string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	var err error
	bpDir, err = filepath.Abs("./..")
	Expect(err).NotTo(HaveOccurred())

	pipURI, err = Package(bpDir, "1.2.3", false)
	Expect(err).ToNot(HaveOccurred())

	pythonURI, err = dagger.GetLatestCommunityBuildpack("paketo-community", "python-runtime")
	Expect(err).ToNot(HaveOccurred())

	defer AfterSuite(t)
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func Package(root, version string, cached bool) (string, error) {
	var cmd *exec.Cmd

	bpPath := filepath.Join(root, "artifact")
	if cached {
		cmd = exec.Command(".bin/packager", "--archive", "--version", version, fmt.Sprintf("%s-cached", bpPath))
	} else {
		cmd = exec.Command(".bin/packager", "--archive", "--uncached", "--version", version, bpPath)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if cached {
		return fmt.Sprintf("%s-cached.tgz", bpPath), err
	}

	return fmt.Sprintf("%s.tgz", bpPath), err
}

func AfterSuite(t *testing.T) {
	var Expect = NewWithT(t).Expect

	Expect(dagger.DeleteBuildpack(pipURI)).To(Succeed())
	Expect(dagger.DeleteBuildpack(pythonURI)).To(Succeed())
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *dagger.App
	)

	it.After(func() {
		Expect(app.Destroy()).To(Succeed())
	})

	when("building a simple app", func() {
		it("runs a python app using pip", func() {
			var err error
			app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app"), pythonURI, pipURI)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				Expect(err).NotTo(HaveOccurred())

				containerID, imageName, volumeIDs, err := app.Info()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.Logs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			body, _, err := app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("Hello, World!"))
		})

		it("caches reused modules for the same app, but downloads new modules ", func() {
			var err error
			app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app"), pythonURI, pipURI)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")
			err = app.Start()
			Expect(err).ToNot(HaveOccurred())

			_, imgName, _, _ := app.Info()

			app, err = dagger.PackBuildNamedImage(imgName, filepath.Join("testdata", "simple_app_more_packages"), pythonURI, pipURI)
			Expect(err).NotTo(HaveOccurred())

			Expect(app.BuildLogs()).To(MatchRegexp("Using cached.*Flask"))
			Expect(app.BuildLogs()).To(MatchRegexp("Downloading.*itsdangerous"))
		})
	})

	when("building a simple app that is vendored", func() {
		it("runs a python app using pip", func() {
			var err error
			app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app"), pythonURI, pipURI)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				Expect(err).NotTo(HaveOccurred())

				containerID, imageName, volumeIDs, err := app.Info()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.Logs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			body, _, err := app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("Hello, World!"))
		})
	})
}
