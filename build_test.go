package pip_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/pip"
	"github.com/paketo-buildpacks/pip/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		cnbDir    string

		dependencyManager  *fakes.DependencyManager
		installProcess     *fakes.InstallProcess
		sitePackageProcess *fakes.SitePackageProcess
		sbomGenerator      *fakes.SBOMGenerator

		logEmitter scribe.Emitter

		buffer *bytes.Buffer

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "pip",
			Name:     "Pip",
			Checksum: "some-sha",
			Stacks:   []string{"some-stack"},
			URI:      "some-uri",
			Version:  "21.0",
		}

		// Legacy SBOM
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "pip",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "pip-dependency-sha",
					},
					URI:     "pip-dependency-uri",
					Version: "pip-dependency-version",
				},
			},
		}

		installProcess = &fakes.InstallProcess{}
		installProcess.ExecuteCall.Stub = func(srcPath, targetLayerPath string) error {
			err = os.MkdirAll(filepath.Join(layersDir, "pip", "lib", "python1.23", "site-packages"), os.ModePerm)
			if err != nil {
				return fmt.Errorf("issue with stub call: %s", err)
			}
			return nil
		}

		sitePackageProcess = &fakes.SitePackageProcess{}
		sitePackageProcess.ExecuteCall.Returns.String = filepath.Join(layersDir, "pip", "lib", "python1.23", "site-packages")

		// Syft SBOM
		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		buffer = bytes.NewBuffer(nil)
		logEmitter = scribe.NewEmitter(buffer)

		build = pip.Build(
			dependencyManager,
			installProcess,
			sitePackageProcess,
			sbomGenerator,
			logEmitter,
			chronos.DefaultClock,
		)

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "pip",
					},
				},
			},
			Platform: packit.Platform{Path: "platform"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs pip", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("pip"))

		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "pip")))

		Expect(layer.SharedEnv).To(HaveLen(2))
		Expect(layer.SharedEnv["PYTHONPATH.delim"]).To(Equal(":"))
		Expect(layer.SharedEnv["PYTHONPATH.prepend"]).To(Equal(filepath.Join(layersDir, "pip", "lib/python1.23/site-packages")))

		Expect(layer.BuildEnv).To(BeEmpty())
		Expect(layer.LaunchEnv).To(BeEmpty())
		Expect(layer.ProcessLaunchEnv).To(BeEmpty())

		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeFalse())
		Expect(layer.Cache).To(BeFalse())

		Expect(layer.Metadata).To(HaveLen(1))
		Expect(layer.Metadata["dependency_checksum"]).To(Equal("some-sha"))

		Expect(layer.SBOM.Formats()).To(Equal([]packit.SBOMFormat{
			{
				Extension: sbom.Format(sbom.CycloneDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.CycloneDXFormat),
			},
			{
				Extension: sbom.Format(sbom.SPDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SPDXFormat),
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("pip"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:       "pip",
			Name:     "Pip",
			Checksum: "some-sha",
			Stacks:   []string{"some-stack"},
			URI:      "some-uri",
			Version:  "21.0",
		}))

		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.DestinationPath).To(ContainSubstring("pip-source"))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "pip")))

		Expect(installProcess.ExecuteCall.Receives.SrcPath).To(Equal(dependencyManager.DeliverCall.Receives.DestinationPath))
		Expect(installProcess.ExecuteCall.Receives.TargetLayerPath).To(Equal(filepath.Join(layersDir, "pip")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing Pip"))
	})

	context("when build plan entries require pip at build/launch", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = make(map[string]interface{})
			buildContext.Plan.Entries[0].Metadata["build"] = true
			buildContext.Plan.Entries[0].Metadata["launch"] = true
		})

		it("makes the layer available at the right times", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("pip"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "pip",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "pip-dependency-sha",
							},
							URI:     "pip-dependency-uri",
							Version: "pip-dependency-version",
						},
					},
				},
			))

			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "pip",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "pip-dependency-sha",
							},
							URI:     "pip-dependency-uri",
							Version: "pip-dependency-version",
						},
					},
				},
			))
		})
	})

	context("when rebuilding a layer", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, fmt.Sprintf("%s.toml", pip.Pip)), []byte(fmt.Sprintf(`[metadata]
			%s = "some-sha"
			built_at = "some-build-time"
			`, pip.DependencyChecksumKey)), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			buildContext.Plan.Entries[0].Metadata = make(map[string]interface{})
			buildContext.Plan.Entries[0].Metadata["build"] = true
			buildContext.Plan.Entries[0].Metadata["launch"] = false
		})

		it("skips the build process if the cached dependency sha matches the selected dependency sha", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("pip"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeFalse())
			Expect(layer.Cache).To(BeTrue())

			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))

			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))
			Expect(installProcess.ExecuteCall.CallCount).To(Equal(0))
		})
	})

	context("failure cases", func() {
		context("when dependency resolution fails", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})
			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("failed to resolve dependency")))
			})
		})

		context("when pip layer cannot be fetched", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when pip layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, pip.Pip), os.ModePerm))
				Expect(os.Chmod(layersDir, 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
			})
			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("failed to install dependency")))
			})
		})

		context("when the site packages cannot be found", func() {
			it.Before(func() {
				sitePackageProcess.ExecuteCall.Returns.Error = errors.New("failed to find site-packages dir")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to find site-packages dir")))
			})
		})

		context("when the layer does not have a site-packages directory", func() {
			it.Before(func() {
				sitePackageProcess.ExecuteCall.Returns.String = ""
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("pip installation failed: site packages are missing from the pip layer")))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})
	})
}
