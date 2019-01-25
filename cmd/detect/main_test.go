package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"
	"path/filepath"
	"pip-cnb/pip"
	"pip-cnb/python_packages"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("there is no requirements.txt", func() {
		it("should fail", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("there is a requirements.txt", func() {
		it.Before(func() {
			Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "requirements.txt"), 0666, "")).To(Succeed())
		})

		when("there is no runtime.txt or buildpack.yml", func() {
			it("should use the default version of python", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())

				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					python.Dependency: buildplan.Dependency{
						Version:  "",
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					pip.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					},
					python_packages.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					},
				}))
			})
		})

		when("there is a runtime.txt but no buildpack.yml", func() {
			const version string = "1.2.3"

			it.Before(func() {
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "requirements.txt"), 0666, "")).To(Succeed())
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "runtime.txt"), 0666, version)).To(Succeed())
			})

			it("should pass with the requested version of python", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())

				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					python.Dependency: buildplan.Dependency{
						Version:  version,
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					pip.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					},
					python_packages.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					},
				}))
			})
		})

		when("there is a buildpack.yml but no runtime.txt", func() {
			const version string = "1.2.3"

			it.Before(func() {
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "requirements.txt"), 0666, "")).To(Succeed())

				buildpackYAMLString := fmt.Sprintf("python:\n  version: %s", version)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of python", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())

				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					python.Dependency: buildplan.Dependency{
						Version:  version,
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					pip.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					},
					python_packages.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					},
				}))
			})
		})

		when("there is a buildpack.yml and a runtime.txt", func() {
			const buildpackYAMLVersion string = "1.2.3"
			const runtimeVersion string = "4.5.6"

			it.Before(func() {
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "requirements.txt"), 0666, "")).To(Succeed())

				buildpackYAMLString := fmt.Sprintf("python:\n  version: %s", buildpackYAMLVersion)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "runtime.txt"), 0666, runtimeVersion)).To(Succeed())
			})

			it("should pass with the requested version of python defined in buildpack.yml", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())

				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					python.Dependency: buildplan.Dependency{
						Version:  buildpackYAMLVersion,
						Metadata: buildplan.Metadata{"build": true, "launch": true},
					},
					pip.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"build": true},
					},
					python_packages.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{"launch": true},
					},
				}))
			})
		})
	})
}
