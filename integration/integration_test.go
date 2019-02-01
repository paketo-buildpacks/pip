package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildpack"

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

			pythonCNBURI, err := dagger.GetRemoteBuildpack("https://github.com/cloudfoundry/python-cnb/releases/download/v0.0.1/python-cnb-0.0.1.tgz")
			Expect(err).ToNot(HaveOccurred())

			builderMetadata := dagger.BuilderMetadata{
				Buildpacks: []dagger.Buildpack{
					{
						ID:  "org.cloudfoundry.buildpacks.python",
						URI: pythonCNBURI,
					},
					{
						ID:  "org.cloudfoundry.buildpacks.pip",
						URI: uri,
					},
				},
				Groups: []dagger.Group{
					{
						[]buildpack.Info{
							{
								ID:      "org.cloudfoundry.buildpacks.python",
								Version: "0.0.1",
							},
							{
								ID:      "org.cloudfoundry.buildpacks.pip",
								Version: "0.0.1",
							},
						},
					},
				},
			}

			app, err := dagger.Pack(filepath.Join("testdata", "simple_app"), builderMetadata, dagger.CFLINUXFS3)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")
			app.Env["PORT"] = "8080"

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				containerID, imageName, volumeIDs, err := app.ContainerInfo()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.ContainerLogs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			err = app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())

			Expect(app.Destroy()).To(Succeed())
		})
	})

	when("building a simple app that is vendored", func() {
		it("runs a python app using pip", func() {
			uri, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			pythonCNBURI, err := dagger.GetRemoteBuildpack("https://github.com/cloudfoundry/python-cnb/releases/download/v0.0.1/python-cnb-0.0.1.tgz")
			Expect(err).ToNot(HaveOccurred())

			builderMetadata := dagger.BuilderMetadata{
				Buildpacks: []dagger.Buildpack{
					{
						ID:  "org.cloudfoundry.buildpacks.python",
						URI: pythonCNBURI,
					},
					{
						ID:  "org.cloudfoundry.buildpacks.pip",
						URI: uri,
					},
				},
				Groups: []dagger.Group{
					{
						[]buildpack.Info{
							{
								ID:      "org.cloudfoundry.buildpacks.python",
								Version: "0.0.1",
							},
							{
								ID:      "org.cloudfoundry.buildpacks.pip",
								Version: "0.0.1",
							},
						},
					},
				},
			}

			app, err := dagger.Pack(filepath.Join("testdata", "simple_app_vendored"), builderMetadata, dagger.CFLINUXFS3)
			Expect(err).ToNot(HaveOccurred())

			app.SetHealthCheck("", "3s", "1s")
			app.Env["PORT"] = "8080"

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
				containerID, imageName, volumeIDs, err := app.ContainerInfo()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("ContainerID: %s\nImage Name: %s\nAll leftover cached volumes: %v\n", containerID, imageName, volumeIDs)

				containerLogs, err := app.ContainerLogs()
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("Container Logs:\n %s\n", containerLogs)
				t.FailNow()
			}

			err = app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())

			Expect(app.Destroy()).To(Succeed())
		})
	})

}
