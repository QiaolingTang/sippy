package resolvedissues

import (
	"fmt"
	"sort"

	"github.com/openshift/sippy/pkg/util/sets"
	"github.com/openshift/sippy/pkg/variantregistry"

	"github.com/openshift/sippy/pkg/apis/api"
)

var triageMatchVariants = buildTriageMatchVariants([]string{variantregistry.VariantPlatform, variantregistry.VariantArch, variantregistry.VariantNetwork,
	variantregistry.VariantTopology, variantregistry.VariantFeatureSet, variantregistry.VariantUpgrade,
	variantregistry.VariantSuite, variantregistry.VariantInstaller})

func buildTriageMatchVariants(in []string) sets.String {
	if in == nil || len(in) < 1 {
		return nil
	}

	set := sets.NewString()

	for _, l := range in {
		set.Insert(l)
	}

	return set
}
func TransformVariant(variant api.ComponentReportColumnIdentification) []api.ComponentReportVariant {
	triagedVariants := []api.ComponentReportVariant{}
	for name, value := range variant.Variants {
		// For now, we only use the defined match variants
		if triageMatchVariants.Has(name) {
			triagedVariants = append(triagedVariants, api.ComponentReportVariant{Key: name, Value: value})
		}
	}
	return triagedVariants
}
func KeyForTriagedIssue(testID string, variants []api.ComponentReportVariant) TriagedIssueKey {

	triagedVariants := make(map[string]string)
	// initialize missing defaults
	triagedVariants[variantregistry.VariantSuite] = "unknown"
	triagedVariants[variantregistry.VariantTopology] = "ha"
	triagedVariants[variantregistry.VariantFeatureSet] = "default"
	triagedVariants[variantregistry.VariantInstaller] = "ipi"

	for _, v := range variants {
		// currently we ignore variants that aren't in api.ComponentReportColumnIdentification
		if triageMatchVariants.Has(v.Key) {
			newValue := v.Value
			switch v.Key {
			case "Upgrade":
				switch v.Value {
				case "upgrade-minor":
					newValue = "minor"
				case "upgrade-micro":
					newValue = "micro"
				case "no-upgrade":
					newValue = "none"
				}
			case "Platform":
				if v.Value == "metal-ipi" {
					newValue = "metal"
				}
			}
			triagedVariants[v.Key] = newValue
		} else if v.Key == "Variant" {
			// inspect the value and create a new key for it to match up with the new variants
			// if the key is part of our triageMatchVariants then add it
			newKey := ""
			newValue := v.Value
			// We have some variant=standard triage records but the discussion was that was just a default value for 'nothing' and isn't needed to
			// be mapped to the new variant standard

			switch v.Value {
			case "proxy":
				newKey = variantregistry.VariantNetworkAccess
			case "fips":
				newKey = variantregistry.VariantSecurityMode
			case "rt":
				newKey = variantregistry.VariantScheduler
				newValue = "realtime"
			case "serial":
				newKey = variantregistry.VariantSuite
			}

			if triageMatchVariants.Has(newKey) {
				triagedVariants[newKey] = newValue
			}
		}
	}

	matchVariants := make([]api.ComponentReportVariant, 0)
	for key, value := range triagedVariants {
		matchVariants = append(matchVariants, api.ComponentReportVariant{Key: key, Value: value})
	}

	sort.Slice(matchVariants,
		func(a, b int) bool {
			return matchVariants[a].Key < matchVariants[b].Key
		})

	vKey := ""
	for _, v := range matchVariants {
		if len(vKey) > 0 {
			vKey += ","
		}
		vKey += fmt.Sprintf("%s_%s", v.Key, v.Value)
	}

	return TriagedIssueKey{
		testID:   testID,
		variants: vKey,
	}
}

type TriageIssueType string

const TriageIssueTypeInfrastructure TriageIssueType = "Infrastructure"

type Release string

type TriagedIssueKey struct {
	testID   string
	variants string
}

type TriagedIncidentsForRelease struct {
	Release          Release                                   `json:"release"`
	TriagedIncidents map[TriagedIssueKey][]api.TriagedIncident `json:"triaged_incidents"`
}

func NewTriagedIncidentsForRelease(release Release) TriagedIncidentsForRelease {
	return TriagedIncidentsForRelease{
		Release:          release,
		TriagedIncidents: map[TriagedIssueKey][]api.TriagedIncident{},
	}
}
