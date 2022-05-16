package pip

import (
	"os"
	"regexp"

	"github.com/paketo-buildpacks/packit/v2"
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
	return func(_ packit.DetectContext) (packit.DetectResult, error) {

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			},
		}

		pipVersion := os.Getenv("BP_PIP_VERSION")

		if pipVersion != "" {
			// Pip releases are of the form X.Y rather than X.Y.0, so in order
			// to support selecting the exact version X.Y we have to up-convert
			// X.Y to X.Y.0.
			// Otherwise X.Y would match the latest patch release
			// X.Y.Z if it is available.
			var xDotYPattern = regexp.MustCompile(`^\d+\.\d+$`)
			if xDotYPattern.MatchString(pipVersion) {
				pipVersion = pipVersion + ".0"
			}

			requirements = append(requirements, packit.BuildPlanRequirement{
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
			},
		}, nil
	}
}
