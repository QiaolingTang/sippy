package regressionallowances

import (
	"encoding/json"
	"fmt"
	"net/url"

	crtype "github.com/openshift/sippy/pkg/apis/api/componentreport"
	"github.com/openshift/sippy/pkg/componentreadiness/resolvedissues"

	log "github.com/sirupsen/logrus"
)

type IntentionalRegression struct {
	JiraComponent             string
	TestID                    string
	TestName                  string
	Variant                   crtype.ColumnIdentification
	PreviousSuccesses         int
	PreviousFailures          int
	PreviousFlakes            int
	RegressedSuccesses        int
	RegressedFailures         int
	RegressedFlakes           int
	JiraBug                   string
	ReasonToAllowInsteadOfFix string
}

type release string

var intentionalRegressions = map[release]map[string]IntentionalRegression{}

type regressionKey struct {
	TestID  string
	Variant crtype.ColumnIdentification
}

func IntentionalRegressionFor(releaseString string, variant crtype.ColumnIdentification, testID string) *IntentionalRegression {
	var targetMap map[string]IntentionalRegression
	var ok bool
	if targetMap, ok = intentionalRegressions[release(releaseString)]; !ok {
		return nil
	}

	inKey := keyFor(testID, variant)
	if t, ok := targetMap[inKey]; ok {
		log.Debugf("found approved regression: %+v", t)
		return &t
	}
	return nil
}

func (i *IntentionalRegression) RegressedPassPercentage(flakeAsFailure bool) float64 {
	return passPercentage(flakeAsFailure, i.RegressedSuccesses, i.RegressedFlakes, i.RegressedFailures)
}

func (i *IntentionalRegression) PreviousPassPercentage(flakeAsFailure bool) float64 {
	return passPercentage(flakeAsFailure, i.PreviousSuccesses, i.PreviousFlakes, i.PreviousFailures)
}

func passPercentage(flakeAsFailure bool, successes, flakes, failures int) float64 {
	if flakeAsFailure {
		return float64(successes) / float64(successes+flakes+failures)
	}
	return float64(successes+flakes) / float64(successes+flakes+failures)
}

func keyFor(testID string, variant crtype.ColumnIdentification) string {
	key := regressionKey{
		TestID: testID,
		Variant: crtype.ColumnIdentification{
			Variants: variant.Variants,
		},
	}
	k, err := json.Marshal(key)
	if err != nil {
		log.WithError(err).Errorf("error marshalling regressionKey")
	}
	return string(k)
}

func parseRegressionKey(key string) (regressionKey, error) {
	var result regressionKey
	if err := json.Unmarshal([]byte(key), &result); err != nil {
		return regressionKey{}, err
	}
	return result, nil
}

func mustAddIntentionalRegression(release release, in IntentionalRegression) {
	if err := addIntentionalRegression(release, in); err != nil {
		panic(err)
	}
}

func addIntentionalRegression(release release, in IntentionalRegression) error {
	if len(in.JiraComponent) == 0 {
		return fmt.Errorf("jiraComponent must be specified")
	}
	if len(in.TestID) == 0 {
		return fmt.Errorf("testID must be specified")
	}
	if len(in.TestName) == 0 {
		return fmt.Errorf("testName must be specified")
	}
	// there must have been successes previously for there to be a regression now
	if in.PreviousSuccesses <= 0 {
		return fmt.Errorf("previousSuccesses must be specified")
	}
	// there must be failures now for there to be a regression
	if in.RegressedFailures <= 0 {
		return fmt.Errorf("regressedFailures must be specified")
	}
	if in.PreviousPassPercentage(false) <= in.RegressedPassPercentage(false) {
		return fmt.Errorf("regressedPassPercentage must be less than previousPassPercentage")
	}
	if len(in.ReasonToAllowInsteadOfFix) == 0 {
		return fmt.Errorf("reasonToAllowInsteadOfFix must be specified")
	}
	if _, err := url.ParseRequestURI(in.JiraBug); err != nil {
		return fmt.Errorf("jiraBug must be a valid URL")
	}
	for _, v := range resolvedissues.TriageMatchVariants.List() {
		if _, ok := in.Variant.Variants[v]; !ok {
			return fmt.Errorf("%s must be specified", v)
		}
	}

	var targetMap map[string]IntentionalRegression
	var ok bool
	if targetMap, ok = intentionalRegressions[release]; !ok {
		targetMap = map[string]IntentionalRegression{}
		intentionalRegressions[release] = targetMap
	}

	inKey := keyFor(in.TestID, in.Variant)
	if _, ok := targetMap[inKey]; ok {
		return fmt.Errorf("test %q was already added", in.TestID)
	}

	targetMap[inKey] = in

	return nil
}
