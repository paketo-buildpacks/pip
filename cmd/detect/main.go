package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-community/pip/python_packages"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
)

const PythonDependency = "python"

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	provided := []buildplan.Provided{
		{Name: python_packages.Dependency},
	}

	exists, err := helper.FileExists(filepath.Join(context.Application.Root, "requirements.txt"))
	if err != nil {
		return detect.FailStatusCode, err
	} else if exists {
		provided = append(provided, buildplan.Provided{Name: python_packages.Requirements})
	}

	requires := []buildplan.Required{
		{
			Name:     PythonDependency,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
		{
			Name:     python_packages.Dependency,
			Metadata: buildplan.Metadata{"launch": true},
		},
		{
			Name:     python_packages.Requirements,
			Metadata: buildplan.Metadata{"build": true},
		},
	}

	return context.Pass(buildplan.Plan{
		Provides: provided,
		Requires: requires,
	})
}
