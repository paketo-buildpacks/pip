package pip

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

// PlanRefinery generates a BuildpackPlan Entry containing the
// Bill-of-Materials of a given dependency.
type PlanRefinery struct{}

// NewPlanRefinery creates a PlanRefinery.
func NewPlanRefinery() PlanRefinery {
	return PlanRefinery{}
}

// BillOfMaterials generates a Bill-of-Materials describing buildpack's
// contributions to the app image.
func (pf PlanRefinery) BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan {
	return packit.BuildpackPlan{
		Entries: []packit.BuildpackPlanEntry{
			{
				Name: dependency.ID,
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     dependency.Name,
					"sha256":   dependency.SHA256,
					"stacks":   dependency.Stacks,
					"uri":      dependency.URI,
					"version":  dependency.Version,
				},
			},
		},
	}
}
