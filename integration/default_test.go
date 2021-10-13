package integration_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("builds with the defaults", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.CPython.Online,
					settings.Buildpacks.Pip.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Pip version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Pip version \(using <unknown>\): \d+\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Pip \d+\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"  Configuring environment",
				MatchRegexp(fmt.Sprintf(`    PYTHONPATH -> "\/layers\/%s\/pip\/lib\/python\d+\.\d+\/site-packages:\$PYTHONPATH"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))

			container, err = docker.Container.Run.
				WithCommand("pip --version").
				Execute(image.ID)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
				// pip versions are not always x.y.z. E.g. 21.3
			}).Should(MatchRegexp(fmt.Sprintf(`pip \d+\.\d+(\.\d+)? from /layers/%s/pip/lib/python\d+.\d+/site-packages/pip`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))))
		})
	})
}
