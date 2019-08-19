package python_packages

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

const (
	Dependency       = "requirements"
	Cache            = "pip_cache"
	RequirementsFile = "requirements.txt"
)

type PackageManager interface {
	Install(requirementsPath, location, cacheDir string) error
	InstallVendor(requirementsPath, location, vendorDir string) error
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

type Contributor struct {
	manager               PackageManager
	app                   application.Application
	packagesLayer         layers.Layer
	packagesLayerMetadata logger.Identifiable
	launchLayer           layers.Layers
	cacheLayer            layers.Layer
	cacheLayerMetadata    logger.Identifiable
	buildContribution     bool
	launchContribution    bool
}

func NewContributor(context build.Build, manager PackageManager) (Contributor, bool, error) {
	plan, willContribute, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil || !willContribute {
		return Contributor{}, false, err
	}

	requirementsFile := filepath.Join(context.Application.Root, RequirementsFile)
	if exists, err := helper.FileExists(requirementsFile); err != nil {
		return Contributor{}, false, err
	} else if !exists {
		return Contributor{}, false, fmt.Errorf(`unable to find "requirements.txt"`)
	}

	contributor := Contributor{
		manager:       manager,
		app:           context.Application,
		packagesLayer: context.Layers.Layer(Dependency),
		cacheLayer:    context.Layers.Layer(Cache),
		launchLayer:   context.Layers,
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	if err := c.contributePythonModules(); err != nil {
		return err
	}

	if err := c.contributePipCache(); err != nil {
		return err
	}

	return c.contributeStartCommand()
}

func (c Contributor) contributePythonModules() error {
	c.packagesLayer.Touch()

	c.packagesLayer.Logger.FirstLine("%s: %s to layer",
		c.packagesLayer.Logger.PrettyIdentity(pythonPackagesID{}), color.YellowString("Contributing"))

	requirements := filepath.Join(c.app.Root, RequirementsFile)
	vendorDir := filepath.Join(c.app.Root, "vendor")

	vendored, err := helper.FileExists(vendorDir)
	if err != nil {
		return fmt.Errorf("unable to stat vendor dir: %s", err.Error())
	}

	if vendored {
		c.packagesLayer.Logger.Info("pip installing from vendor directory")
		if err := c.manager.InstallVendor(requirements, c.packagesLayer.Root, vendorDir); err != nil {
			return err
		}
	} else {
		c.packagesLayer.Logger.Info("pip installing to: " + c.packagesLayer.Root)
		if err := c.manager.Install(requirements, c.packagesLayer.Root, c.cacheLayer.Root); err != nil {
			return err
		}
	}

	if err := c.packagesLayer.AppendPathSharedEnv("PYTHONUSERBASE", c.packagesLayer.Root); err != nil {
		return err
	}

	return c.packagesLayer.WriteMetadata(nil, c.flags()...)
}

func (c Contributor) contributeStartCommand() error {
	procfile := filepath.Join(c.app.Root, "Procfile")
	exists, err := helper.FileExists(procfile)
	if err != nil {
		return err
	}

	if exists {
		buf, err := ioutil.ReadFile(procfile)
		if err != nil {
			return err
		}

		proc := regexp.MustCompile(`^\s*web\s*:\s*`).ReplaceAllString(string(buf), "")
		return c.launchLayer.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{Type: "web", Command: proc}}})
	}

	return nil
}

func (c Contributor) contributePipCache() error {
	if cacheExists, err := helper.FileExists(c.cacheLayer.Root); err != nil {
		return err
	} else if cacheExists {
		c.cacheLayer.Touch()

		c.cacheLayer.Logger.FirstLine("%s: %s to layer",
			c.cacheLayer.Logger.PrettyIdentity(pipCacheID{}), color.YellowString("Contributing"))

		return c.cacheLayer.WriteMetadata(nil, layers.Cache)
	}
	return nil
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}
	return flags
}

type pythonPackagesID struct {
}

func (p pythonPackagesID) Identity() (name string, description string) {
	return "Python Packages", "latest"
}

type pipCacheID struct {
}

func (p pipCacheID) Identity() (name string, description string) {
	return "PIP Cache", "latest"
}
