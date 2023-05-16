package pip_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/pip"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect        packit.DetectFunc
		detectContext packit.DetectContext
	)

	it.Before(func() {
		detect = pip.Detect()
		detectContext = packit.DetectContext{}
	})

	context("detection", func() {
		it("returns a build plan that provides pip", func() {
			result, err := detect(detectContext)
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
				result, err := detect(detectContext)
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
				}))
			})

			context("when the provided version is of the form X.Y", func() {
				it.Before(func() {
					os.Setenv("BP_PIP_VERSION", "2.11")
				})

				it("selects the version X.Y.0", func() {
					result, err := detect(detectContext)

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
									Version:       "2.11.0",
									VersionSource: "BP_PIP_VERSION",
								},
							},
						},
					}))
				})
			})

			context("when the provided version is of the form X.Y.Z", func() {
				it.Before(func() {
					os.Setenv("BP_PIP_VERSION", "22.1.3")
				})

				it("selects the exact provided version X.Y.Z", func() {
					result, err := detect(detectContext)

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
									Version:       "22.1.3",
									VersionSource: "BP_PIP_VERSION",
								},
							},
						},
					}))
				})
			})

			context("when the provided version is of some other form", func() {
				it.Before(func() {
					os.Setenv("BP_PIP_VERSION", "some.other")
				})

				it("selects the exact provided version", func() {
					result, err := detect(detectContext)

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
									Version:       "some.other",
									VersionSource: "BP_PIP_VERSION",
								},
							},
						},
					}))
				})
			})
		})

	})
}
