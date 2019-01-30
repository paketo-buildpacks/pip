package python_packages

import (
	"fmt"
	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

const (
	Dependency       = "python_packages"
	RequirementsFile = "requirements.txt"
)

type PackageManager interface {
	Install(requirementsPath, location string) error
	InstallVendor(requirementsPath, location, vendorDir string) error
}

type Contributor struct {
	manager            PackageManager
	app                application.Application
	packagesLayer      layers.Layer
	launchLayer        layers.Layers
	buildContribution  bool
	launchContribution bool
}

func NewContributor(context build.Build, manager PackageManager) (Contributor, bool, error) {
	dep, willContribute := context.BuildPlan[Dependency]
	if !willContribute {
		return Contributor{}, false, nil
	}

	contributor := Contributor{
		manager:       manager,
		app:           context.Application,
		packagesLayer: context.Layers.Layer(Dependency),
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


// if -> we have the same metadata (sha of our python_modules)
// then -> we should re-install all of this stuff.
// also we might want to look at how pip packages are
func (c Contributor) Contribute() error {
	return c.packagesLayer.Contribute(nil, func(layer layers.Layer) error {
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
			if err := c.manager.Install(requirements, c.packagesLayer.Root); err != nil {
				return err
			}
		}

		if err := layer.AppendPathSharedEnv("PYTHONPATH", c.packagesLayer.Root); err != nil {
			return err
		}

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
	}, c.flags()...)
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
