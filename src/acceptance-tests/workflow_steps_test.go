package acceptance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cucumber/godog"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
)

type workflowScenarioState struct {
	mux                   *http.ServeMux
	mockSrv               *httptest.Server
	client                *github.Client
	receivedCreatedFilter string
	runsResult            []github.WorkflowRun
	jobsResult            []github.Job
	withJobsResult        []github.RunWithJobs
	fetchErr              error
}

func (s *workflowScenarioState) writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *workflowScenarioState) runJSON(id int64, name string) map[string]interface{} {
	return map[string]interface{}{
		"id": id, "name": name, "head_branch": "main",
		"event": "push", "status": "completed", "conclusion": "success",
		"created_at": "2026-06-15T10:00:00Z", "updated_at": "2026-06-15T10:05:00Z",
		"actor": map[string]interface{}{"login": "alice"},
	}
}

func (s *workflowScenarioState) runsPageJSON(runs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"total_count": len(runs), "workflow_runs": runs}
}

func (s *workflowScenarioState) jobJSON(id int64, name, conclusion string) map[string]interface{} {
	return map[string]interface{}{
		"id": id, "name": name, "status": "completed", "conclusion": conclusion,
	}
}

func (s *workflowScenarioState) jobsPageJSON(jobs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"total_count": len(jobs), "jobs": jobs}
}

// aRepositoryWithWorkflowRunsAvailable registers a single-page mock returning 2 runs.
// The handler captures the "created" query parameter so the no-filter assertion can check it.
func (s *workflowScenarioState) aRepositoryWithWorkflowRunsAvailable() error {
	s.mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		s.receivedCreatedFilter = r.URL.Query().Get("created")
		s.writeJSON(w, http.StatusOK, s.runsPageJSON([]map[string]interface{}{
			s.runJSON(1, "CI"), s.runJSON(2, "CI"),
		}))
	})
	return nil
}

func (s *workflowScenarioState) theExporterFetchesWorkflowRunsForTheRepository() error {
	s.runsResult, s.fetchErr = s.client.WorkflowRuns(context.Background(), "owner", "repo")
	return nil
}

func (s *workflowScenarioState) theMostRecentWorkflowRunsAreReturned() error {
	if len(s.runsResult) == 0 {
		return fmt.Errorf("expected workflow runs to be returned, got none")
	}
	return nil
}

func (s *workflowScenarioState) noTimeWindowFilterIsAppliedToTheGitHubAPIRequest() error {
	if s.receivedCreatedFilter != "" {
		return fmt.Errorf("expected no time-window filter, but got created=%q", s.receivedCreatedFilter)
	}
	return nil
}

// aWorkflowRunWith3Jobs registers a mock returning 3 jobs for run 1.
func (s *workflowScenarioState) aWorkflowRunWith3Jobs() error {
	s.mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.jobsPageJSON([]map[string]interface{}{
			s.jobJSON(10, "build", "success"),
			s.jobJSON(11, "test", "failure"),
			s.jobJSON(12, "lint", "success"),
		}))
	})
	return nil
}

func (s *workflowScenarioState) theExporterFetchesJobsForThatRun() error {
	s.jobsResult, s.fetchErr = s.client.JobsForRun(context.Background(), "owner", "repo", 1)
	return nil
}

func (s *workflowScenarioState) allJobsAreReturnedWithTheirNameAndConclusion(expectedCount int) error {
	if len(s.jobsResult) != expectedCount {
		return fmt.Errorf("expected %d jobs, got %d", expectedCount, len(s.jobsResult))
	}
	for i, j := range s.jobsResult {
		if j.Name == "" || j.Conclusion == "" {
			return fmt.Errorf("job %d missing name or conclusion: name=%q conclusion=%q", i, j.Name, j.Conclusion)
		}
	}
	return nil
}

// aRepositoryWith2WorkflowRunsWhereTheFirstRunHasAFailingJobsEndpoint sets up:
// - runs endpoint returning 2 runs (IDs 1 and 2)
// - run 1 jobs endpoint returns 500
// - run 2 jobs endpoint returns 1 successful job
func (s *workflowScenarioState) aRepositoryWith2WorkflowRunsWhereTheFirstRunHasAFailingJobsEndpoint() error {
	s.mux.HandleFunc("/repos/owner/repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.runsPageJSON([]map[string]interface{}{
			s.runJSON(1, "CI"), s.runJSON(2, "CI"),
		}))
	})
	s.mux.HandleFunc("/repos/owner/repo/actions/runs/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	})
	s.mux.HandleFunc("/repos/owner/repo/actions/runs/2/jobs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.jobsPageJSON([]map[string]interface{}{
			s.jobJSON(20, "deploy", "success"),
		}))
	})
	return nil
}

func (s *workflowScenarioState) theExporterFetchesWorkflowsWithJobsForTheRepository() error {
	s.withJobsResult, s.fetchErr = s.client.FetchWorkflowsWithJobs(context.Background(), "owner", "repo")
	return nil
}

func (s *workflowScenarioState) theFetchDoesNotReturnATopLevelError() error {
	if s.fetchErr != nil {
		return fmt.Errorf("expected no top-level error, got %v", s.fetchErr)
	}
	return nil
}

func (s *workflowScenarioState) bothRunsArePresentInTheResult() error {
	if len(s.withJobsResult) != 2 {
		return fmt.Errorf("expected 2 runs in result, got %d", len(s.withJobsResult))
	}
	return nil
}

func (s *workflowScenarioState) theRunWithTheJobErrorHasAnEmptyJobList() error {
	for _, r := range s.withJobsResult {
		if r.Run.ID == 1 {
			if len(r.Jobs) != 0 {
				return fmt.Errorf("expected empty jobs for run 1 (job fetch failed), got %d jobs", len(r.Jobs))
			}
			return nil
		}
	}
	return fmt.Errorf("run with ID 1 not found in result")
}

func (s *workflowScenarioState) scrapingContinuesAndTheSecondRunHasItsJobs() error {
	for _, r := range s.withJobsResult {
		if r.Run.ID == 2 {
			if len(r.Jobs) == 0 {
				return fmt.Errorf("expected run 2 to have jobs, got none")
			}
			return nil
		}
	}
	return fmt.Errorf("run with ID 2 not found in result")
}

// InitializeWorkflowScenario registers workflow step definitions with GoDog.
func InitializeWorkflowScenario(ctx *godog.ScenarioContext) {
	s := &workflowScenarioState{}

	ctx.Before(func(goCtx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.mux = http.NewServeMux()
		s.mockSrv = httptest.NewServer(s.mux)
		s.client = github.NewWithToken("test-token")
		u, err := url.Parse(s.mockSrv.URL + "/")
		if err != nil {
			return goCtx, fmt.Errorf("setup: parse mock server URL: %w", err)
		}
		s.client.SetBaseURL(u)
		s.receivedCreatedFilter = ""
		s.runsResult = nil
		s.jobsResult = nil
		s.withJobsResult = nil
		s.fetchErr = nil
		return goCtx, nil
	})

	ctx.After(func(goCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if s.mockSrv != nil {
			s.mockSrv.Close()
		}
		return goCtx, nil
	})

	ctx.Step(`^a repository with workflow runs available from the GitHub API$`, s.aRepositoryWithWorkflowRunsAvailable)
	ctx.Step(`^the exporter fetches workflow runs for the repository$`, s.theExporterFetchesWorkflowRunsForTheRepository)
	ctx.Step(`^the most recent workflow runs are returned$`, s.theMostRecentWorkflowRunsAreReturned)
	ctx.Step(`^no time-window filter is applied to the GitHub API request$`, s.noTimeWindowFilterIsAppliedToTheGitHubAPIRequest)
	ctx.Step(`^a workflow run with 3 jobs$`, s.aWorkflowRunWith3Jobs)
	ctx.Step(`^the exporter fetches jobs for that run$`, s.theExporterFetchesJobsForThatRun)
	ctx.Step(`^all (\d+) jobs are returned with their name and conclusion$`, s.allJobsAreReturnedWithTheirNameAndConclusion)
	ctx.Step(`^a repository with 2 workflow runs where the first run has a failing jobs endpoint$`, s.aRepositoryWith2WorkflowRunsWhereTheFirstRunHasAFailingJobsEndpoint)
	ctx.Step(`^the exporter fetches workflows with jobs for the repository$`, s.theExporterFetchesWorkflowsWithJobsForTheRepository)
	ctx.Step(`^the fetch does not return a top-level error$`, s.theFetchDoesNotReturnATopLevelError)
	ctx.Step(`^both runs are present in the result$`, s.bothRunsArePresentInTheResult)
	ctx.Step(`^the run with the job error has an empty job list$`, s.theRunWithTheJobErrorHasAnEmptyJobList)
	ctx.Step(`^scraping continues and the second run has its jobs$`, s.scrapingContinuesAndTheSecondRunHasItsJobs)
}
