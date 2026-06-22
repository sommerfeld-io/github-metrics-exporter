Feature: Workflow run and job data retrieval
  When the GitHub Metrics Exporter is running,
  I want workflow and job data fetched from the GitHub API for each discovered repository
  So that workflow-level and job-level metrics can be calculated

  Scenario: Most recent workflow runs are fetched for a repository
    Given a repository with workflow runs available from the GitHub API
    When the exporter fetches workflow runs for the repository
    Then the most recent workflow runs are returned
    And no time-window filter is applied to the GitHub API request

  Scenario: Jobs retrieved for a workflow run
    Given a workflow run with 3 jobs
    When the exporter fetches jobs for that run
    Then all 3 jobs are returned with their name and conclusion

  Scenario: GitHub API returns an error for one run's jobs endpoint
    Given a repository with 2 workflow runs where the first run has a failing jobs endpoint
    When the exporter fetches workflows with jobs for the repository
    Then the fetch does not return a top-level error
    And both runs are present in the result
    And the run with the job error has an empty job list
    And scraping continues and the second run has its jobs

  Scenario Outline: Workflow run conclusion appears as a metric on the metrics endpoint
    Given the exporter has repositories configured
    When a user requests the metrics endpoint
    Then the metrics body contains workflow run conclusion "<conclusion>" for "test-org" and "repo-accessible"

    Examples:
      | conclusion |
      | success    |
      | failure    |

  Scenario: Workflow job conclusion appears as a metric on the metrics endpoint
    Given the exporter has repositories configured
    When a user requests the metrics endpoint
    Then the metrics body contains workflow job conclusion "success" for "test-org" and "repo-accessible"
