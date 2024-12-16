package prowloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeNameComparisons(t *testing.T) {
	tests := []struct {
		name           string
		names          []string
		expectedResult string
	}{
		{
			name: "standard",
			names: []string{`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`,
				`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-153052.json`},
			expectedResult: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`,
		},
		{
			name: "reversed",
			names: []string{`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-153052.json`,
				`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`},
			expectedResult: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`,
		},
		{
			name: "older date",
			names: []string{`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230219-153052.json`,
				`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`},
			expectedResult: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230219-153052.json`,
		},
		{
			name: "invalid date",
			names: []string{`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230219153052.json`,
				`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`},
			expectedResult: `https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230218-180228.json`,
		},
		{
			name: "invalid dates",
			names: []string{`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/cluster-data_20230219153052.json`,
				`https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/27731/pull-ci-openshift-origin-master-e2e-aws-ovn-upgrade/1626951434970861568/artifacts/e2e-aws-ovn-upgrade/openshift-e2e-test/artifacts/junit/.json`},
			expectedResult: ``,
		},
		{
			name:           "empty names",
			names:          []string{},
			expectedResult: ``,
		},
		{
			name:           "no names",
			expectedResult: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, findMostRecentDateTimeMatch(tt.names), "Test: %s failed mostRecentDateTimeMatch", tt.name)
		})
	}
}

func TestParseVariantDataFile(t *testing.T) {
	clusterDataFile := []byte(`{
    "Release": "4.16",
    "FromRelease": "4.15",
    "Platform": "gcp",
    "Architecture": "amd64",
    "Network": "ovn",
    "Topology": "ha",
    "NetworkStack": "IPv4",
    "CloudRegion": "us-central1",
    "CloudZone": "us-central1-a",
    "AddonProp1": "foo",
    "ClusterVersionHistory": [
        "4.16.0-0.nightly-2024-02-21-020511",
        "4.15.0-0.nightly-2024-02-20-090411"
    ],
    "MasterNodesUpdated": "Y"
}`)
	clusterData, err := ParseVariantDataFile(clusterDataFile)
	assert.NoError(t, err)
	assert.Equal(t, "4.16", clusterData["Release"])
	assert.Equal(t, "4.15", clusterData["FromRelease"])
	assert.Equal(t, "gcp", clusterData["Platform"])
	assert.Equal(t, "IPv4", clusterData["NetworkStack"])
	assert.Equal(t, "foo", clusterData["AddonProp1"])
}
