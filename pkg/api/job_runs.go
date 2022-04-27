package api

import (
	"encoding/json"
	"net/http"
	gosort "sort"
	"strconv"
	"strings"

	apitype "github.com/openshift/sippy/pkg/apis/api"
	v1sippyprocessing "github.com/openshift/sippy/pkg/apis/sippyprocessing/v1"
	"github.com/openshift/sippy/pkg/db"
	"github.com/openshift/sippy/pkg/filter"
	"github.com/openshift/sippy/pkg/util"
)

func (runs apiRunResults) sort(req *http.Request) apiRunResults {
	sortField := req.URL.Query().Get("sortField")
	sort := apitype.Sort(req.URL.Query().Get("sort"))

	if sortField == "" {
		sortField = "test_failures"
	}

	if sort == "" {
		sort = apitype.SortDescending
	}

	gosort.Slice(runs, func(i, j int) bool {
		if sort == apitype.SortAscending {
			return filter.Compare(runs[i], runs[j], sortField)
		}
		return filter.Compare(runs[j], runs[i], sortField)
	})

	return runs
}

func (runs apiRunResults) limit(req *http.Request) apiRunResults {
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	if limit > 0 && len(runs) >= limit {
		return runs[:limit]
	}

	return runs
}

type apiRunResults []apitype.JobRun

func jobRunToAPIJobRun(id int, job v1sippyprocessing.JobResult, result v1sippyprocessing.JobRunResult) apitype.JobRun {
	return apitype.JobRun{
		ID:                    id,
		BriefName:             briefName(job.Name),
		Variants:              job.Variants,
		TestGridURL:           job.TestGridURL,
		ProwID:                result.ProwID,
		Job:                   result.Job,
		URL:                   result.URL,
		TestFailures:          result.TestFailures,
		FailedTestNames:       result.FailedTestNames,
		Failed:                result.Failed,
		InfrastructureFailure: result.InfrastructureFailure,
		KnownFailure:          result.KnownFailure,
		Succeeded:             result.Succeeded,
		Timestamp:             result.Timestamp,
		OverallResult:         result.OverallResult,
	}
}

// PrintJobRunsReport renders the detailed list of runs for matching jobs.
func PrintJobRunsReport(w http.ResponseWriter, req *http.Request, currReport, prevReport v1sippyprocessing.TestReport) {
	var fil *filter.Filter
	curr := currReport.ByJob
	prev := prevReport.ByJob

	queryFilter := req.URL.Query().Get("filter")
	if queryFilter != "" {
		fil = &filter.Filter{}
		if err := json.Unmarshal([]byte(queryFilter), fil); err != nil {
			RespondWithJSON(http.StatusBadRequest, w, map[string]interface{}{"code": http.StatusBadRequest, "message": "Could not marshal query:" + err.Error()})
			return
		}
	}

	all := make([]apitype.JobRun, 0)
	next := 0
	for _, results := range append(curr, prev...) {
		for _, run := range results.AllRuns {
			apiRun := jobRunToAPIJobRun(next, results, run)

			if fil != nil {
				include, err := fil.Filter(apiRun)

				// Job runs are a little special, in that we want to let users filter them by fields from the job
				// itself, too.
				if err != nil && strings.Contains(err.Error(), "unknown") {
					currJob := util.FindJobResultForJobName(run.Job, curr)
					if currJob != nil {
						prevJob := util.FindJobResultForJobName(run.Job, prev)
						include, err = fil.Filter(jobResultToAPI(next, currJob, prevJob))
					}
				}
				if err != nil {
					RespondWithJSON(http.StatusBadRequest, w, map[string]interface{}{"code": http.StatusBadRequest, "message": "Filter error:" + err.Error()})
					return
				}

				if !include {
					continue
				}
			}

			all = append(all, apiRun)
			next++
		}
	}

	RespondWithJSON(http.StatusOK, w,
		apiRunResults(all).
			sort(req).
			limit(req),
	)
}

// PrintJobsRunsReportFromDB renders a filtered summary of matching jobs.
func PrintJobsRunsReportFromDB(w http.ResponseWriter, req *http.Request,
	dbc *db.DB, release string) {

	var fil *filter.Filter

	queryFilter := req.URL.Query().Get("filter")
	if queryFilter != "" {
		fil = &filter.Filter{}
		if err := json.Unmarshal([]byte(queryFilter), fil); err != nil {
			RespondWithJSON(http.StatusBadRequest, w, map[string]interface{}{"code": http.StatusBadRequest, "message": "Could not marshal query:" + err.Error()})
			return
		}
	}

	filterOpts, err := filter.FilterOptionsFromRequest(req, "timestamp", "desc")
	if err != nil {
		RespondWithJSON(http.StatusInternalServerError, w, map[string]interface{}{"code": http.StatusInternalServerError, "message": "Error building job run report:" + err.Error()})
		return
	}

	rf := releaseFilter(req, dbc.DB)
	q, err := filter.FilterableDBResult(dbc.DB, filterOpts, apitype.JobRun{})
	if err != nil {
		RespondWithJSON(http.StatusInternalServerError, w, map[string]interface{}{"code": http.StatusInternalServerError, "message": "Error building job run report:" + err.Error()})
		return
	}
	q = rf.Where(q)

	jobsResult := make([]apitype.JobRun, 0)
	q.Table("prow_job_runs_report_matview").Scan(&jobsResult)
	if err != nil {
		RespondWithJSON(http.StatusInternalServerError, w, map[string]interface{}{"code": http.StatusInternalServerError, "message": "Error building job report:" + err.Error()})
		return
	}

	RespondWithJSON(http.StatusOK, w, jobsResult)
}
