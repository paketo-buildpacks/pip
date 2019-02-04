package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("building a simple app", func() {
		it("runs a python app using pip", func() {
			uri, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			pythonCNBURI, err := dagger.GetLatestBuildpack("python-cnb")
			Expect(err).ToNot(HaveOccurred())

			app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), pythonCNBURI, uri)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")
			app.Env["PORT"] = "8080"

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
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

			Expect(app.Destroy()).To(Succeed())
		})
	})

	when("building a simple app that is vendored", func() {
		it("runs a python app using pip", func() {
			uri, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			pythonCNBURI, err := dagger.GetLatestBuildpack("python-cnb")
			Expect(err).ToNot(HaveOccurred())

			app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), pythonCNBURI, uri)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")
			app.Env["PORT"] = "8080"

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
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

			Expect(app.Destroy()).To(Succeed())
		})
	})
}
