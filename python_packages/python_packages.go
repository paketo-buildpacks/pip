package python_packages

import (
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
	PackagesDir      = "packages"
	RequirementsFile = "requirements.txt"
)

type PackageManager interface {
	Install(requirementsPath, location string) error
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

func (c Contributor) Contribute() error {
	return c.packagesLayer.Contribute(nil, func(layer layers.Layer) error {
		requirements := filepath.Join(c.app.Root, RequirementsFile)
		packages := filepath.Join(c.packagesLayer.Root, PackagesDir)

		if err := c.manager.Install(requirements, packages); err != nil {
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
