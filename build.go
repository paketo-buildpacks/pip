package pip

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackageProcess --output fakes/site_package_process.go

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

// DependencyManager defines the interface for picking the best matching
// dependency and installing it.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, destPath string) error
}

// BuildPlanRefinery defines the interface for generating a BuildpackPlan Entry
// containing the Bill-of-Materials of a given dependency.
type BuildPlanRefinery interface {
	BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan
}

// InstallProcess defines the interface for installing the pip dependency into a layer.
type InstallProcess interface {
	Execute(srcPath, targetLayerPath string) error
}

// SitePackageProcess defines the interface for looking site packages within a layer.
type SitePackageProcess interface {
	Execute(targetLayerPath string) (string, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right pip dependency to install, install it in a
// layer, and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(installProcess InstallProcess, entries EntryResolver, dependencies DependencyManager, planRefinery BuildPlanRefinery, logs scribe.Emitter, clock chronos.Clock, siteProcess SitePackageProcess) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logs.Process("Resolving Pip version")

		entry, sortedEntries := entries.Resolve(Pip, context.Plan.Entries, Priorities)

		logs.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		dependency.Name = "Pip"
		logs.SelectedDependency(entry, dependency, clock.Now())

		bom := planRefinery.BillOfMaterial(dependency)

		pipLayer, err := context.Layers.Get(Pip)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := pipLayer.Metadata[DependencySHAKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", pipLayer.Path)
			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{pipLayer},
			}, nil
		}

		pipLayer, err = pipLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipLayer.Launch, pipLayer.Build = entries.MergeLayerTypes(Pip, context.Plan.Entries)
		pipLayer.Cache = pipLayer.Build

		// Install the pip source to a temporary dir, since we only need access to
		// it as an intermediate step when installing pip.
		// It doesn't need to go into a layer, since we won't need it in future builds.
		pipSrcDir, err := ioutil.TempDir("", "pip-source")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to create temp pip-source dir: %w", err)
		}

		logs.Process("Executing build process")
		logs.Subprocess(fmt.Sprintf("Installing Pip %s", dependency.Version))

		duration, err := clock.Measure(func() error {
			err = dependencies.Install(dependency, context.CNBPath, pipSrcDir)
			if err != nil {
				return err
			}
			return installProcess.Execute(pipSrcDir, pipLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.Action("Completed in %s", duration.Round(time.Millisecond))
		logs.Break()

		// Look up the site packages path and prepend it onto $PYTHONPATH
		sitePackagesPath, err := siteProcess.Execute(pipLayer.Path)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to locate site packages in pip layer: %w", err)
		}
		if sitePackagesPath == "" {
			return packit.BuildResult{}, fmt.Errorf("pip installation failed: site packages are missing from the pip layer")
		}

		pipLayer.SharedEnv.Prepend("PYTHONPATH", strings.TrimRight(sitePackagesPath, "\n"), ":")

		logs.Process("Configuring environment")
		logs.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(pipLayer.SharedEnv))
		logs.Break()

		pipLayer.Metadata = map[string]interface{}{
			DependencySHAKey: dependency.SHA256,
			"built_at":       clock.Now().Format(time.RFC3339Nano),
		}

		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{pipLayer},
		}, nil
	}
}
