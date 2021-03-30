package pip

import (
	"os"

	"github.com/paketo-buildpacks/packit"
)

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	// Build denotes the dependency is needed at build-time.
	Build bool `toml:"build"`

	// Launch denotes the dependency is needed at runtime.
	Launch bool `toml:"launch"`

	// Version denotes the version of a dependency, if there is one.
	Version string `toml:"version"`

	// VersionSource denotes where dependency version came from (e.g. an environment variable).
	VersionSource string `toml:"version-source"`
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection always passes, and will contribute a  Build Plan that provides pip,
// and requires cpython OR python, python_packages, and requirements.
//
// If a version is provided via the $BP_PIP_VERSION environment variable, that
// version of pip will be a requirement.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
		}

		// TODO: remove this when removing legacy API
		requirementsLegacy := []packit.BuildPlanRequirement{
			{
				Name: Python,
				Metadata: BuildPlanMetadata{
					Build:  true,
					Launch: true,
				},
			},
			{
				Name: Requirements,
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
		}

		pipVersion := os.Getenv("BP_PIP_VERSION")

		if pipVersion != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Pip,
				Metadata: BuildPlanMetadata{
					VersionSource: "BP_PIP_VERSION",
					Version:       pipVersion,
				},
			})

			// TODO: remove this when removing legacy API
			requirementsLegacy = append(requirementsLegacy, packit.BuildPlanRequirement{
				Name: Pip,
				Metadata: BuildPlanMetadata{
					VersionSource: "BP_PIP_VERSION",
					Version:       pipVersion,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Pip},
				},
				Requires: requirements,
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: Pip},
						},
						Requires: requirementsLegacy,
					},
				},
			},
		}, nil
	}
}
