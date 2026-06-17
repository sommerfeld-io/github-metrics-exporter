# Package: `config`

Loads and validates the exporter's YAML configuration file. The `Load` function reads the file at a given path, enforces strict type checking on all fields, and returns a populated `Config` struct or a descriptive error that identifies exactly which field failed and why. Any error causes the caller to abort startup before the HTTP server is started.
