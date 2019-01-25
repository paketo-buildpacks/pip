package pip

import (
	"fmt"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"io/ioutil"
	"os"
	"os/exec"
)

const Dependency = "pip"

type Contributor struct {
	pipLayer layers.DependencyLayer
	buildContribution bool
	launchContribution bool
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, willContribute := context.BuildPlan[Dependency]
	if willContribute == false {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		pipLayer: context.Layers.DependencyLayer(dep),
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, willContribute, nil
}

func (c Contributor) Contribute() error {
	return c.pipLayer.Contribute(c.contributePipLayer, c.flags()...)
}

func (c Contributor) contributePipLayer(artifact string, layer layers.DependencyLayer) error {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err := helper.ExtractTarGz(artifact, tmp, 1); err != nil {
		return err
	}

	cmd := exec.Command("python", "setup.py", "install", fmt.Sprintf("--prefix=%s", layer.Root))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmp
	return cmd.Run()
}

func (n Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
