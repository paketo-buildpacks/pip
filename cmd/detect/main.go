package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"pip-cnb/python_packages"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"
	"gopkg.in/yaml.v2"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default detect context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	if err := context.BuildPlan.Init(); err != nil {
		return detect.FailStatusCode, err
	}

	if willContribute, err := willContribute(context); err != nil {
		return detect.FailStatusCode, err
	} else if !willContribute {
		return detect.FailStatusCode, nil
	}

	runtimePath := filepath.Join(context.Application.Root, "runtime.txt")
	exists, err := helper.FileExists(runtimePath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	var version string
	if exists {
		buf, err := ioutil.ReadFile(runtimePath)
		if err != nil {
			return detect.FailStatusCode, err
		}
		version = string(buf)
	}

	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err = helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	if exists {
		buf, err := ioutil.ReadFile(buildpackYAMLPath)
		if err != nil {
			return detect.FailStatusCode, err
		}

		config := struct {
			Python struct {
				Version string `yaml:"version"`
			} `yaml:"python"`
		}{}
		if err := yaml.Unmarshal(buf, &config); err != nil {
			return detect.FailStatusCode, err
		}

		version = config.Python.Version
	}

	return context.Pass(buildplan.BuildPlan{
		python.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
		python_packages.Dependency: buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}

// TODO: Refactor to a detector package
func willContribute(context detect.Detect) (bool, error) {
	_, ok := context.BuildPlan[python_packages.Dependency]

	if ok {
		context.Logger.Info("pip packages requested by previous buildpack")
		return true, nil
	}

	if exists, err := helper.FileExists(filepath.Join(context.Application.Root, "requirements.txt")); err != nil {
		return false, err
	} else if !exists {
		context.Logger.Info("no requirements.txt found")
		return false, nil
	}

	return true, nil
}
