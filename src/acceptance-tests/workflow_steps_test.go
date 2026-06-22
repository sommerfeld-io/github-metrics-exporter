package acceptance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cucumber/godog"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/github"
)

type workflowScenarioState struct {
	mux              *http.ServeMux
	mockSrv          *httptest.Server
	client           *github.Client
	originalLogger   *slog.Logger
	workflowsResult  []github.Workflow
	jobsResult       []github.Job
	withJobsResult   []github.RunWithJobs
	fetchErr         error
}

func (s *workflowScenarioState) writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *workflowScenarioState) workflowDefJSON(id int64, name, path string) map[string]interface{} {
	return map[string]interface{}{"id": id, "name": name, "path": path}
}

func (s *workflowScenarioState) workflowsPageJSON(wfs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"total_count": len(wfs), "workflows": wfs}
}

func (s *workflowScenarioState) runPageJSON(id int64, name string) map[string]interface{} {
	run := map[string]interface{}{
		"id": id, "name": name, "head_branch": "main",
		"event": "push", "status": "completed", "conclusion": "success",
		"created_at": "2026-06-15T10:00:00Z", "updated_at": "2026-06-15T10:05:00Z",
		"actor": map[string]interface{}{"login": "alice"},
	}
	return map[string]interface{}{"total_count": 1, "workflow_runs": []interface{}{run}}
}

func (s *workflowScenarioState) jobJSON(id int64, name, conclusion string) map[string]interface{} {
	return map[string]interface{}{
		"id": id, "name": name, "status": "completed", "conclusion": conclusion,
	}
}

func (s *workflowScenarioState) jobsPageJSON(jobs []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"total_count": len(jobs), "jobs": jobs}
}

// aRepositoryWithWorkflowDefinitionsAvailable registers a mock returning 2 workflow definitions.
func (s *workflowScenarioState) aRepositoryWithWorkflowDefinitionsAvailable() error {
	s.mux.HandleFunc("/repos/owner/repo/actions/workflows", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.workflowsPageJSON([]map[string]interface{}{
			s.workflowDefJSON(1, "CI", ".github/workflows/ci.yml"),
			s.workflowDefJSON(2, "Release", ".github/workflows/release.yml"),
		}))
	})
	return nil
}

func (s *workflowScenarioState) theExporterListsWorkflowDefinitionsForTheRepository() error {
	s.workflowsResult, s.fetchErr = s.client.ListWorkflows(context.Background(), "owner", "repo")
	return nil
}

func (s *workflowScenarioState) allWorkflowDefinitionsAreReturned() error {
	if len(s.workflowsResult) == 0 {
		return fmt.Errorf("expected workflow definitions to be returned, got none")
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

// aRepositoryWith2WorkflowsWhereTheFirstRunHasAFailingJobsEndpoint sets up:
//   - workflows endpoint returning 2 workflow definitions (IDs 1 and 2)
//   - per-workflow run endpoints returning run ID 1 and run ID 2 respectively
//   - run 1 jobs endpoint returns 500
//   - run 2 jobs endpoint returns 1 successful job
func (s *workflowScenarioState) aRepositoryWith2WorkflowsWhereTheFirstRunHasAFailingJobsEndpoint() error {
	s.mux.HandleFunc("/repos/owner/repo/actions/workflows", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.workflowsPageJSON([]map[string]interface{}{
			s.workflowDefJSON(1, "CI", ".github/workflows/ci.yml"),
			s.workflowDefJSON(2, "Release", ".github/workflows/release.yml"),
		}))
	})
	s.mux.HandleFunc("/repos/owner/repo/actions/workflows/1/runs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.runPageJSON(1, "CI"))
	})
	s.mux.HandleFunc("/repos/owner/repo/actions/workflows/2/runs", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, s.runPageJSON(2, "Release"))
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
		s.originalLogger = slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		s.mux = http.NewServeMux()
		s.mockSrv = httptest.NewServer(s.mux)
		s.client = github.NewWithToken("test-token")
		u, err := url.Parse(s.mockSrv.URL + "/")
		if err != nil {
			return goCtx, fmt.Errorf("setup: parse mock server URL: %w", err)
		}
		s.client.SetBaseURL(u)
		s.workflowsResult = nil
		s.jobsResult = nil
		s.withJobsResult = nil
		s.fetchErr = nil
		return goCtx, nil
	})

	ctx.After(func(goCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		slog.SetDefault(s.originalLogger)
		if s.mockSrv != nil {
			s.mockSrv.Close()
		}
		return goCtx, nil
	})

	ctx.Step(`^a repository with workflow definitions available from the GitHub API$`, s.aRepositoryWithWorkflowDefinitionsAvailable)
	ctx.Step(`^the exporter lists workflow definitions for the repository$`, s.theExporterListsWorkflowDefinitionsForTheRepository)
	ctx.Step(`^all workflow definitions are returned$`, s.allWorkflowDefinitionsAreReturned)
	ctx.Step(`^a workflow run with 3 jobs$`, s.aWorkflowRunWith3Jobs)
	ctx.Step(`^the exporter fetches jobs for that run$`, s.theExporterFetchesJobsForThatRun)
	ctx.Step(`^all (\d+) jobs are returned with their name and conclusion$`, s.allJobsAreReturnedWithTheirNameAndConclusion)
	ctx.Step(`^a repository with 2 workflows where the first run has a failing jobs endpoint$`, s.aRepositoryWith2WorkflowsWhereTheFirstRunHasAFailingJobsEndpoint)
	ctx.Step(`^the exporter fetches workflows with jobs for the repository$`, s.theExporterFetchesWorkflowsWithJobsForTheRepository)
	ctx.Step(`^the fetch does not return a top-level error$`, s.theFetchDoesNotReturnATopLevelError)
	ctx.Step(`^both runs are present in the result$`, s.bothRunsArePresentInTheResult)
	ctx.Step(`^the run with the job error has an empty job list$`, s.theRunWithTheJobErrorHasAnEmptyJobList)
	ctx.Step(`^scraping continues and the second run has its jobs$`, s.scrapingContinuesAndTheSecondRunHasItsJobs)
}
