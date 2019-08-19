package main

import (
	"testing"

	"github.com/cloudfoundry/pip-cnb/python_packages"

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

	when("there is no requirements.txt", func() {
		it("passes and requires that dependency", func() {
			code, err := runDetect(factory.Detect)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(getStandardBuildplan([]buildplan.Provided{})))
		})
	})

	when("the app has a requirements.txt", func() {
		it("requires and provides that dependency", func() {
			test.TouchFile(t, factory.Detect.Application.Root, "requirements.txt")
			code, err := runDetect(factory.Detect)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))

			Expect(factory.Plans.Plan).To(Equal(getStandardBuildplan([]buildplan.Provided{{Name: python_packages.Dependency}})))
		})
	})

}

func getStandardBuildplan(provides []buildplan.Provided) buildplan.Plan {
	return buildplan.Plan{
		Provides: provides,
		Requires: []buildplan.Required{
			{
				Name: python.Dependency,
				Metadata: buildplan.Metadata{
					"build":  true,
					"launch": true,
				},
			},
			{
				Name:     python_packages.Dependency,
				Metadata: buildplan.Metadata{"launch": true},
			},
		},
	}
}
