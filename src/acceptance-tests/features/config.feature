Feature: Startup Configuration Validation
  When the GitHub Metrics Exporter is running,
  I want the exporter to validate the configuration file at startup
  So that misconfiguration is caught immediately with a descriptive error

  Scenario Outline: Invalid port types are rejected at startup
    Given a config file where port is <yaml_value>
    When the config is loaded
    Then the config loader returns an error

    Examples:
      | yaml_value |
      | "9400"     |
      | 9400.5     |
      | true       |
      | ~          |

  Scenario: Missing port key is rejected at startup
    Given a config file with no port key
    When the config is loaded
    Then the config loader returns an error

  Scenario Outline: Valid integer port is accepted at startup
    Given a config file where port is <port>
    When the config is loaded
    Then the config loader succeeds with port <port>

    Examples:
      | port |
      | 9400 |
      | 8080 |
