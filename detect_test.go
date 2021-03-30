package pip_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-community/pip"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		detect packit.DetectFunc
	)

	it.Before(func() {
		detect = pip.Detect()
	})

	context("detection", func() {
		it("returns a build plan that provides pip", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: pip.Pip},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: pip.CPython,
						Metadata: pip.BuildPlanMetadata{
							Build: true,
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: pip.Pip},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: pip.Python,
								Metadata: pip.BuildPlanMetadata{
									Build:  true,
									Launch: true,
								},
							},
							{
								Name: pip.Requirements,
								Metadata: pip.BuildPlanMetadata{
									Build: true,
								},
							},
						},
					},
				},
			}))
		})

		context("when BP_PIP_VERSION is set", func() {
			it.Before(func() {
				os.Setenv("BP_PIP_VERSION", "some-version")
			})

			it.After(func() {
				os.Unsetenv("BP_PIP_VERSION")
			})

			it("returns a build plan that provides the version of pip from BP_PIP_VERSION", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: pip.Pip},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pip.CPython,
							Metadata: pip.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Pip,
							Metadata: pip.BuildPlanMetadata{
								Version:       "some-version",
								VersionSource: "BP_PIP_VERSION",
							},
						},
					},
					Or: []packit.BuildPlan{
						{
							Provides: []packit.BuildPlanProvision{
								{Name: pip.Pip},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name: pip.Python,
									Metadata: pip.BuildPlanMetadata{
										Build:  true,
										Launch: true,
									},
								},
								{
									Name: pip.Requirements,
									Metadata: pip.BuildPlanMetadata{
										Build: true,
									},
								},
								{
									Name: pip.Pip,
									Metadata: pip.BuildPlanMetadata{
										Version:       "some-version",
										VersionSource: "BP_PIP_VERSION",
									},
								},
							},
						},
					},
				}))
			})
		})

	})
}
