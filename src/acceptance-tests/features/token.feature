Feature: GitHub Token validation at startup
  When the GitHub Metrics Exporter is running,
  I want the exporter to validate the GitHub token at startup
  So that configuration errors are detected early and do not cause silent failures at scrape time

  Scenario: Token is missing at startup
    Given the exporter is started without GITHUB_TOKEN set
    When the startup sequence validates the token
    Then the validation returns a clear error indicating the missing token
    And the error message mentions "GITHUB_TOKEN"

  Scenario: Token environment variable is set but empty
    Given the exporter is started with GITHUB_TOKEN set to an empty string
    When the startup sequence validates the token
    Then the validation returns a clear error indicating the missing token
    And the error message mentions "GITHUB_TOKEN"

  Scenario: Token is present at startup
    Given the exporter is started with a non-empty GITHUB_TOKEN
    When the startup sequence validates the token
    Then the validation succeeds without error
