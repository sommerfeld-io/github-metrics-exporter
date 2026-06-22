Feature: Observability of metrics scrape requests
  When a Prometheus scraper or operator hits /metrics,
  I want the exporter to log the start and end of each data-collection cycle
  So that activity is visible in the application logs without opening the metrics output itself

  Scenario: A log entry is written when the metrics scrape starts
    Given the exporter has repositories configured
    When a user requests the metrics endpoint
    Then a log entry with message "metrics: scrape started" is written

  Scenario: A log entry is written when the metrics scrape completes
    Given the exporter has repositories configured
    When a user requests the metrics endpoint
    Then a log entry with message "metrics: scrape completed" is written
