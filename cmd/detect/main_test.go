package main

import (
	"pip-cnb/python_packages"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/python-cnb/python"

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

	when("there is no requirements.txt and no buildplan", func() {
		it("should fail detection", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("python packages are requested in the buildplan", func() {
		it("passes", func() {
			factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{})
			factory.AddBuildPlan(python.Dependency, buildplan.Dependency{
				Version:  "",
				Metadata: buildplan.Metadata{"build": true, "launch": true},
			})

			code, err := runDetect(factory.Detect)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))

			Expect(factory.Output).To(Equal(buildplan.BuildPlan{
				python.Dependency: buildplan.Dependency{
					Version:  "",
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				},
				python_packages.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				},
			}))
		})
	})

}
