package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/pip-cnb/python_packages"

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

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	exists, err := helper.FileExists(filepath.Join(context.Application.Root, "requirements.txt"))
	if err != nil {
		return detect.FailStatusCode, err
	}

	provided := []buildplan.Provided{}
	if exists {
		provided = append(provided, buildplan.Provided{Name: python_packages.Dependency})
	}

	return context.Pass(buildplan.Plan{
		Provides: provided,
		Requires: []buildplan.Required{
			{
				Name:     python.Dependency,
				Metadata: buildplan.Metadata{"build": true, "launch": true},
			},
			{
				Name:     python_packages.Dependency,
				Metadata: buildplan.Metadata{"launch": true},
			},
		},
	})
}
