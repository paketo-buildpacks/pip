package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"os"
	"pip-cnb/pip"
	"pip-cnb/python_packages"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default build context: %s", err)
		os.Exit(100)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	packagesContributor, willContribute, err := python_packages.NewContributor(context, pip.PIP{})
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := packagesContributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildplan.BuildPlan{})
}
