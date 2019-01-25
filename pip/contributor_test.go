package pip_test

import (
	"github.com/sclevine/spec/report"
	"path/filepath"
	"pip-cnb/pip"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitPipContributor(t *testing.T) {
	spec.Run(t, "PipContributor", testContributor, spec.Report(report.Terminal{}))
}

func testContributor(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("NewContributor", func() {
		var stubPipFixture = filepath.Join("testdata", "stub-pip.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(pip.Dependency, buildplan.Dependency{})
			f.AddDependency(pip.Dependency, stubPipFixture)

			_, willContribute, err := pip.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := pip.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("contributes pip to the cache layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(pip.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true},
			})
			f.AddDependency(pip.Dependency, stubPipFixture)

			pipDep, _, err := pip.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(pipDep.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(pip.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})

		it("contributes pip to the launch layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(pip.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(pip.Dependency, stubPipFixture)

			pipContributor, _, err := pip.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(pipContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(pip.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})
	})
}
