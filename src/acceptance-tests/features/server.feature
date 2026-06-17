Feature: Serve Dark-Themed Landing Page, Prometheus Metrics, and Core Exporter Meta-Metrics
  When the GitHub Metrics Exporter is running,
  I want the root endpoint ("/") to serve a dark-themed informational page with project links and build metadata, and the "/metrics" endpoint to expose standard Prometheus metrics alongside a set of core exporter meta-info metrics (such as the build commit SHA),
  so that I can easily verify the operational state of the exporter manually while allowing Grafana Alloy to scrape both performance data and structural version metadata.

  Scenario Outline: Verifying HTTP response headers and status routing
    Given the exporter application is running and healthy
    When a user requests the path "<path>"
    Then the HTTP status code should be <status_code>
    And the response content-type should contain "<content_type>"

    Examples:
      | path           | status_code | content_type |
      | /              | 200         | text/html    |
      | /metrics       | 200         | text/plain   |
      | /healthz       | 200         | text/plain   |
      | /invalid-route | 404         | text/plain   |
      | /wrong/path    | 404         | text/plain   |

  Scenario: Verifying the root endpoint HTML content layout and theme
    Given the exporter application is running and healthy
    When a user requests the path "/"
    Then the page should contain the headline "GitHub Metrics Exporter"
    And the page should contain a link to "https://github.com/sommerfeld-io/github-metrics-exporter/"
    And the page should contain a link to "/metrics"
    And the page should display the static build commit SHA
    And the page source HTML style should declare a dark background theme
    And the page should display the configured port

  Scenario: Verifying default core meta-metrics presence
    Given the exporter application is running and healthy
    When a user requests the path "/metrics"
    Then the response body must contain the default metric "ghme_exporter_info" with a commit_sha label
