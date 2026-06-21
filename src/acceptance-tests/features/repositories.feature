Feature: Repository discovery and accessibility
  As a platform engineer
  I want repositories to be discovered automatically from configured GitHub targets
  So that onboarding new repositories requires no manual configuration

  Scenario Outline: Repository accessibility state is exposed as a metric
    Given the exporter has repositories configured
    When a user requests the metrics endpoint
    Then the metrics body contains a repository accessible metric with value <value> for "<owner>" and "<repo>"

    Examples:
      | owner    | repo            | value |
      | test-org | repo-accessible | 1     |
      | test-org | repo-locked     | 0     |

  Scenario: Repository section appears on the index page grouped by owner
    Given the exporter has repositories configured
    When a user navigates to "/"
    Then the page should contain the owner heading "test-org"
    And the page should list repository "repo-accessible"
    And the page should list repository "repo-locked"

  Scenario Outline: Repository accessibility is indicated with a badge on the index page
    Given the exporter has repositories configured
    When a user navigates to "/"
    Then the page should show an "<badge>" badge for "<repo>"

    Examples:
      | repo            | badge        |
      | repo-accessible | accessible   |
      | repo-locked     | inaccessible |

  Scenario: Warning appears on index page when no targets are configured
    Given the exporter has no targets configured
    When a user navigates to "/"
    Then the page should display the no-targets warning
