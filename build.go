package pip

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackageProcess --output fakes/site_package_process.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// DependencyManager defines the interface for picking the best matching
// dependency and installing it.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, destinationPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// InstallProcess defines the interface for installing the pip dependency into a layer.
type InstallProcess interface {
	Execute(srcPath, targetLayerPath string) error
}

// SitePackageProcess defines the interface for looking site packages within a layer.
type SitePackageProcess interface {
	Execute(targetLayerPath string) (string, error)
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right pip dependency to install, install it in a
// layer, and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	dependencies DependencyManager,
	installProcess InstallProcess,
	siteProcess SitePackageProcess,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		logger.Process("Resolving Pip version")
		entry, sortedEntries := planner.Resolve(Pip, context.Plan.Entries, Priorities)
		logger.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		dependency.Name = "Pip"
		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes(Pip, context.Plan.Entries)

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = legacySBOM
		}

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = legacySBOM
		}

		pipLayer, err := context.Layers.Get(Pip)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := pipLayer.Metadata[DependencySHAKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", pipLayer.Path)
			pipLayer.Launch, pipLayer.Build, pipLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{pipLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		pipLayer, err = pipLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipLayer.Launch, pipLayer.Build, pipLayer.Cache = launch, build, build

		// Install the pip source to a temporary dir, since we only need access to
		// it as an intermediate step when installing pip.
		// It doesn't need to go into a layer, since we won't need it in future builds.
		pipSrcDir, err := os.MkdirTemp("", "pip-source")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to create temp pip-source dir: %w", err)
		}

		logger.Process("Executing build process")
		logger.Subprocess(fmt.Sprintf("Installing Pip %s", dependency.Version))

		duration, err := clock.Measure(func() error {
			err = dependencies.Deliver(dependency, context.CNBPath, pipSrcDir, context.Platform.Path)
			if err != nil {
				return err
			}
			return installProcess.Execute(pipSrcDir, pipLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(pipLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, pipLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		pipLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// Look up the site packages path and prepend it onto $PYTHONPATH
		sitePackagesPath, err := siteProcess.Execute(pipLayer.Path)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to locate site packages in pip layer: %w", err)
		}
		if sitePackagesPath == "" {
			return packit.BuildResult{}, fmt.Errorf("pip installation failed: site packages are missing from the pip layer")
		}
		pipLayer.SharedEnv.Prepend("PYTHONPATH", strings.TrimRight(sitePackagesPath, "\n"), ":")

		logger.EnvironmentVariables(pipLayer)

		pipLayer.Metadata = map[string]interface{}{
			DependencySHAKey: dependency.SHA256,
		}

		return packit.BuildResult{
			Layers: []packit.Layer{pipLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
