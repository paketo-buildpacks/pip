package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/pip"
)

func main() {
	entries := draft.NewPlanner()
	dependencies := postal.NewService(cargo.NewTransport())
	logs := scribe.NewEmitter(os.Stdout)
	installProcess := pip.NewPipInstallProcess(pexec.NewExecutable("python"))
	siteProcess := pip.NewSiteProcess(pexec.NewExecutable("python"))

	packit.Run(
		pip.Detect(),
		pip.Build(installProcess, entries, dependencies, logs, chronos.DefaultClock, siteProcess),
	)
}
