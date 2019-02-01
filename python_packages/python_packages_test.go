package python_packages_test

import (
	"os"
	"path/filepath"
	"pip-cnb/python_packages"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=python_packages.go -destination=mocks_test.go -package=python_packages_test

func TestUnitPythonPackages(t *testing.T) {
	spec.Run(t, "PythonPackages", testPythonPackages, spec.Report(report.Terminal{}))
}

func testPythonPackages(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("modules.NewContributor", func() {
		var (
			mockCtrl       *gomock.Controller
			mockPkgManager *MockPackageManager
			factory        *test.BuildFactory
		)

		it.Before(func() {
			mockCtrl = gomock.NewController(t)
			mockPkgManager = NewMockPackageManager(mockCtrl)

			factory = test.NewBuildFactory(t)
		})

		it.After(func() {
			mockCtrl.Finish()
		})

		it("NewContributor returns willContribute true if a build plan exists with the dep", func() {
			test.TouchFile(t, filepath.Join(factory.Build.Application.Root, "requirements.txt"))

			factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{})

			_, willContribute, err := python_packages.NewContributor(factory.Build, mockPkgManager)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("NewContributor returns willContribute false if a build plan does not exist with the dep", func() {
			_, willContribute, err := python_packages.NewContributor(factory.Build, mockPkgManager)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		when("the app is vendored", func() {
			var vendorDir, vendorPackage string
			it.Before(func() {
				requirementsPath := filepath.Join(factory.Build.Application.Root, "requirements.txt")
				test.TouchFile(t, requirementsPath)

				vendorDir = filepath.Join(factory.Build.Application.Root, "vendor")
				os.MkdirAll(vendorDir, 0777)

				packages := factory.Build.Layers.Layer(python_packages.Dependency).Root
				vendorPackage = filepath.Join(packages, "vendoredFile")

				mockPkgManager.EXPECT().InstallVendor(requirementsPath, packages, vendorDir).Do(func(_, packages, _ string) {
					Expect(os.MkdirAll(packages, os.ModePerm)).To(Succeed())
					test.WriteFile(t, vendorPackage, "vendored package contents")
				})
			})
			it.After(func() {
				os.Remove(vendorPackage)
				os.RemoveAll(vendorDir)
			})
			it("contributes for the build phase", func() {
				factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{
					Metadata: buildplan.Metadata{"build": true},
				})

				contributor, _, err := python_packages.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(contributor.Contribute()).To(Succeed())

				packagesLayer := factory.Build.Layers.Layer(python_packages.Dependency)
				Expect(packagesLayer).To(test.HaveLayerMetadata(true, true, false))

				Expect(filepath.Join(packagesLayer.Root, "vendoredFile")).To(BeARegularFile())
			})

			it("contributes for the launch phase", func() {
				procFileString := "web: gunicorn server:app"
				Expect(helper.WriteFile(filepath.Join(factory.Build.Application.Root, "Procfile"), 0666, procFileString)).To(Succeed())

				factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true},
				})

				contributor, _, err := python_packages.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(contributor.Contribute()).To(Succeed())

				Expect(factory.Build.Layers).To(test.HaveLaunchMetadata(layers.Metadata{Processes: []layers.Process{{"web", "gunicorn server:app"}}}))
				packagesLayer := factory.Build.Layers.Layer(python_packages.Dependency)
				Expect(packagesLayer).To(test.HaveLayerMetadata(false, true, true))
				Expect(filepath.Join(packagesLayer.Root, "vendoredFile")).To(BeARegularFile())
			})
		})

		when("the app is not vendored", func() {
			it.Before(func() {
				requirementsPath := filepath.Join(factory.Build.Application.Root, "requirements.txt")
				test.TouchFile(t, requirementsPath)
				packages := factory.Build.Layers.Layer(python_packages.Dependency).Root
				cacheDir := factory.Build.Layers.Layer(python_packages.Cache).Root

				mockPkgManager.EXPECT().Install(requirementsPath, packages, cacheDir).Do(func(_, packages, _ string) {
					Expect(os.MkdirAll(packages, os.ModePerm)).To(Succeed())
					test.WriteFile(t, filepath.Join(packages, "package"), "package contents")

					Expect(os.MkdirAll(cacheDir, os.ModePerm)).To(Succeed())
				})
			})

			it("contributes for the build phase", func() {
				factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{
					Metadata: buildplan.Metadata{"build": true},
				})

				contributor, _, err := python_packages.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(contributor.Contribute()).To(Succeed())

				packagesLayer := factory.Build.Layers.Layer(python_packages.Dependency)
				cacheLayer := factory.Build.Layers.Layer(python_packages.Cache)
				Expect(packagesLayer).To(test.HaveLayerMetadata(true, true, false))
				Expect(cacheLayer).To(test.HaveLayerMetadata(false, true, false))
				Expect(filepath.Join(packagesLayer.Root, "package")).To(BeARegularFile())
			})

			it("contributes for the launch phase", func() {
				procFileString := "web: gunicorn server:app"
				Expect(helper.WriteFile(filepath.Join(factory.Build.Application.Root, "Procfile"), 0666, procFileString)).To(Succeed())

				factory.AddBuildPlan(python_packages.Dependency, buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true},
				})

				contributor, _, err := python_packages.NewContributor(factory.Build, mockPkgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(contributor.Contribute()).To(Succeed())

				Expect(factory.Build.Layers).To(test.HaveLaunchMetadata(layers.Metadata{Processes: []layers.Process{{"web", "gunicorn server:app"}}}))

				packagesLayer := factory.Build.Layers.Layer(python_packages.Dependency)
				cacheLayer := factory.Build.Layers.Layer(python_packages.Cache)
				Expect(packagesLayer).To(test.HaveLayerMetadata(false, true, true))
				Expect(cacheLayer).To(test.HaveLayerMetadata(false, true, false))
				Expect(filepath.Join(packagesLayer.Root, "package")).To(BeARegularFile())
			})
		})
	})
}
