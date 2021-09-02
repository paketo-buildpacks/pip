package pip_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	pip "github.com/paketo-community/pip"
	"github.com/paketo-community/pip/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir          string
		cnbDir             string
		entryResolver      *fakes.EntryResolver
		dependencyManager  *fakes.DependencyManager
		clock              chronos.Clock
		timeStamp          time.Time
		installProcess     *fakes.InstallProcess
		sitePackageProcess *fakes.SitePackageProcess
		buffer             *bytes.Buffer
		logEmitter         scribe.Emitter

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
  id = "org.some-org.some-buildpack"
  name = "Some Buildpack"
  version = "some-version"

[metadata]

  [[metadata.dependencies]]
		id = "pip"
    name = "Pip"
    sha256 = "some-sha"
    stacks = ["some-stack"]
    uri = "some-uri"
    version = "21.0"
`), 0600)
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "pip",
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "pip",
			Name:    "Pip",
			SHA256:  "some-sha",
			Stacks:  []string{"some-stack"},
			URI:     "some-uri",
			Version: "21.0",
		}
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "pip",
				Metadata: packit.BOMMetadata{
					Checksum: packit.BOMChecksum{
						Algorithm: packit.SHA256,
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

		buffer = bytes.NewBuffer(nil)
		logEmitter = scribe.NewEmitter(buffer)

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		build = pip.Build(installProcess, entryResolver, dependencyManager, logEmitter, clock, sitePackageProcess)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs pip", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "pip",
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
			Stack:  "some-stack",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name: "pip",
					Path: filepath.Join(layersDir, "pip"),
					SharedEnv: packit.Environment{
						"PYTHONPATH.delim":   ":",
						"PYTHONPATH.prepend": filepath.Join(layersDir, "pip", "lib/python1.23/site-packages"),
					},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					ProcessLaunchEnv: map[string]packit.Environment{},
					Metadata: map[string]interface{}{
						pip.DependencySHAKey: "some-sha",
						"built_at":           timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.String).To(Equal("pip"))
		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "pip",
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.InterfaceSlice).To(Equal([]interface{}{"BP_PIP_VERSION"}))

		Expect(entryResolver.MergeLayerTypesCall.Receives.String).To(Equal("pip"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.BuildpackPlanEntrySlice).To(Equal(
			[]packit.BuildpackPlanEntry{
				{
					Name: "pip",
				},
			},
		))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("pip"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "pip",
			Name:    "Pip",
			SHA256:  "some-sha",
			Stacks:  []string{"some-stack"},
			URI:     "some-uri",
			Version: "21.0",
		}))

		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.DestPath).To(ContainSubstring("pip-source"))

		Expect(installProcess.ExecuteCall.Receives.SrcPath).To(Equal(dependencyManager.InstallCall.Receives.DestPath))
		Expect(installProcess.ExecuteCall.Receives.TargetLayerPath).To(Equal(filepath.Join(layersDir, "pip")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing Pip"))
		Expect(buffer.String()).To(ContainSubstring("Configuring environment"))
	})

	context("when there's an entry with version source BP_PIP_VERSION", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Build = true
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
		})

		it("the BP_PIP_VERSION version takes precedence", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "pip",
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
						{
							Name: "pip",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name: "pip",
						Path: filepath.Join(layersDir, "pip"),
						SharedEnv: packit.Environment{
							"PYTHONPATH.delim":   ":",
							"PYTHONPATH.prepend": filepath.Join(layersDir, "pip", "lib/python1.23/site-packages"),
						},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						ProcessLaunchEnv: map[string]packit.Environment{},
						Metadata: map[string]interface{}{
							pip.DependencySHAKey: "some-sha",
							"built_at":           timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "pip",
							Metadata: packit.BOMMetadata{
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "pip-dependency-sha",
								},
								URI:     "pip-dependency-uri",
								Version: "pip-dependency-version",
							},
						},
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "pip",
							Metadata: packit.BOMMetadata{
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "pip-dependency-sha",
								},
								URI:     "pip-dependency-uri",
								Version: "pip-dependency-version",
							},
						},
					},
				},
			}))
		})
	})
	context("when build plan entries require pip at build/launch", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Build = true
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
		})

		it("makes the layer available at the right times", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "pip",
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
						{
							Name: "pip",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(Equal([]packit.Layer{
				{
					Name: "pip",
					Path: filepath.Join(layersDir, "pip"),
					SharedEnv: packit.Environment{
						"PYTHONPATH.delim":   ":",
						"PYTHONPATH.prepend": filepath.Join(layersDir, "pip", "lib/python1.23/site-packages"),
					},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					Build:            true,
					Launch:           true,
					Cache:            true,
					ProcessLaunchEnv: map[string]packit.Environment{},
					Metadata: map[string]interface{}{
						pip.DependencySHAKey: "some-sha",
						"built_at":           timeStamp.Format(time.RFC3339Nano),
					},
				},
			}))
		})
	})

	context("when rebuilding a layer", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(layersDir, fmt.Sprintf("%s.toml", pip.Pip)), []byte(fmt.Sprintf(`[metadata]
			%s = "some-sha"
			built_at = "some-build-time"
			`, pip.DependencySHAKey)), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(layersDir, "pip", "env"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(layersDir, "pip", "env", "PYTHONPATH.prepend"), []byte(fmt.Sprintf("%s/pip/lib/python1.23/site-packages", layersDir)), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(layersDir, "pip", "env", "PYTHONPATH.delim"), []byte(":"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		it("skips the build process if the cached dependency sha matches the selected dependency sha", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "pip",
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))

			Expect(result.Layers).To(Equal([]packit.Layer{
				{
					Name: "pip",
					Path: filepath.Join(layersDir, "pip"),
					SharedEnv: packit.Environment{
						"PYTHONPATH.delim":   ":",
						"PYTHONPATH.prepend": filepath.Join(layersDir, "pip", "lib/python1.23/site-packages"),
					},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					ProcessLaunchEnv: map[string]packit.Environment{},
					Metadata: map[string]interface{}{
						pip.DependencySHAKey: "some-sha",
						"built_at":           "some-build-time",
					},
				},
			}))

			Expect(dependencyManager.InstallCall.CallCount).To(Equal(0))
			Expect(installProcess.ExecuteCall.CallCount).To(Equal(0))
		})
	})

	context("failure cases", func() {
		context("when dependency resolution fails", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})

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
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})

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
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})

				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})

				Expect(err).To(MatchError(ContainSubstring("failed to install dependency")))
			})
		})

		context("when the site packages cannot be found", func() {
			it.Before(func() {
				sitePackageProcess.ExecuteCall.Returns.Error = errors.New("failed to find site-packages dir")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to find site-packages dir")))
			})
		})

		context("when the layer does not have a site-packages directory", func() {
			it.Before(func() {
				sitePackageProcess.ExecuteCall.Returns.String = ""
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "pip",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("pip installation failed: site packages are missing from the pip layer")))
			})
		})

	})
}
