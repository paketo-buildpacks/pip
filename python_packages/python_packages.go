package python_packages

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency       = "python_packages"
	Cache            = "pip_cache"
	RequirementsFile = "requirements.txt"
)

type PackageManager interface {
	Install(requirementsPath, location, cacheDir string) error
	InstallVendor(requirementsPath, location, vendorDir string) error
}

type Contributor struct {
	manager            PackageManager
	app                application.Application
	packagesLayer      layers.Layer
	launchLayer        layers.Layers
	cacheLayer         layers.Layer
	buildContribution  bool
	launchContribution bool
}

func NewContributor(context build.Build, manager PackageManager) (Contributor, bool, error) {
	dep, willContribute := context.BuildPlan[Dependency]
	if !willContribute {
		return Contributor{}, false, nil
	}

	requirementsFile := filepath.Join(context.Application.Root, "requirements.txt")
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

	if _, ok := dep.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := dep.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {

	if err := c.packagesLayer.Contribute(nil, c.contributePythonModules, c.flags()...); err != nil {
		return err
	}
	return c.cacheLayer.Contribute(nil, c.contributePipCache, layers.Cache)
}

// cache will have been written to layer during contributePythonModules so we just add to layer
func (c Contributor) contributePipCache(layer layers.Layer) error {
	if err := os.MkdirAll(layer.Root, 0777); err != nil {
		return fmt.Errorf("unable make pip cache layer: %s", err.Error())
	}

	if empty, err := isEmptyDir(layer.Root); err != nil || empty {
		if err != nil {
			layer.Logger.Info("cache dir does not exist")
		} else {
			layer.Logger.Info("Did not contribute cache layer")
		}
	} else {
		layer.Logger.Info("contributed cache layer")
	}
	return nil
}

func (c Contributor) contributePythonModules(layer layers.Layer) error {
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

	if err := layer.AppendPathSharedEnv("PYTHONUSERBASE", c.packagesLayer.Root); err != nil {
		return err
	}

	return c.contributeStartCommand(layer)
}

func (c Contributor) contributeStartCommand(layer layers.Layer) error {
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
		return c.launchLayer.WriteMetadata(layers.Metadata{Processes: []layers.Process{{"web", proc}}})
	}
	return nil
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}
	return flags
}

func isEmptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
