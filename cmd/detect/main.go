package main

import (
	"fmt"
	"os"
	"path/filepath"
	"pip-cnb/python_packages"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	if err := context.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(101)
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

	version := context.BuildPlan[python.Dependency].Version

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
