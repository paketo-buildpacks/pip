package pip

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
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

		pipSrcLayer, err := context.Layers.Get(PipSrc)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedChecksum, ok := pipLayer.Metadata[DependencyChecksumKey].(string)
		if ok && cargo.Checksum(cachedChecksum).Match(cargo.Checksum(dependency.Checksum)) {
			logger.Process("Reusing cached layer %s", pipLayer.Path)
			logger.Process("Reusing cached layer %s", pipSrcLayer.Path)
			pipLayer.Launch, pipLayer.Build, pipLayer.Cache = launch, build, build
			pipSrcLayer.Launch, pipSrcLayer.Build, pipSrcLayer.Cache = false, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{pipLayer, pipSrcLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		pipLayer, err = pipLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipSrcLayer, err = pipSrcLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipLayer.Launch, pipLayer.Build, pipLayer.Cache = launch, build, build
		//Pip-source layer flags should mirror the Pip layer, but should never be
		//available at launch.
		pipSrcLayer.Launch, pipSrcLayer.Build, pipSrcLayer.Cache = false, build, build

		logger.Process("Executing build process")
		logger.Subprocess(fmt.Sprintf("Installing Pip %s", dependency.Version))

		duration, err := clock.Measure(func() error {
			err = dependencies.Deliver(dependency, context.CNBPath, pipSrcLayer.Path, context.Platform.Path)
			if err != nil {
				return err
			}
			return installProcess.Execute(pipSrcLayer.Path, pipLayer.Path)
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

		// Append the pip source layer path to PIP_FIND_LINKS so that invocations
		// of pip in downstream buildpacks have access to the packages bundled with
		// the pip dependency (setuptools, wheel, etc.).

		pipSrcLayer.BuildEnv.Append("PIP_FIND_LINKS", strings.TrimRight(pipSrcLayer.Path, "\n"), " ")

		logger.EnvironmentVariables(pipSrcLayer)
		logger.EnvironmentVariables(pipLayer)

		pipLayer.Metadata = map[string]interface{}{
			DependencyChecksumKey: dependency.Checksum,
		}

		return packit.BuildResult{
			Layers: []packit.Layer{pipLayer, pipSrcLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
