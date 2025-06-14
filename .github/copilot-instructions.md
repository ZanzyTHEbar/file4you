
Do not manually manage dependancies, write the code and let go manage the deps. Proceed with our real work.

- make sure to update all documentation, todos, and reports as you complete each task
- make sure to prefix all placeholder logic with `// TODO: ` to indicate incomplete work
- make sure to regularly run when relevant `go fmt ./...` to format all code
- make sure to regularly run when relevant `go vet ./...` to check for any issues
- make sure to regularly run when relevant `go test ./...` to ensure all tests pass
- make sure to regularly run when relevant `go build ./...` to ensure all code compiles
- make sure to implement robust table-driven tests for all core functionality

## Required Packages

Do NOT cycle on manually managing dependancies and ALWAYS focus on our high-value-add work. Dependancies are managed by go and do not need to be manually managed.

DO NOT waste time on dependency management and focus on high-value work.

If `go mod tidy` is ran before the dependencies are used, it will remove the dependencies that are not used. Therefore, make sure to add AND USE _all_ required dependencies before running `go mod tidy`.

- make sure to use the latest version of all dependencies
- make sure to use the following packages:
  - `github.com/spf13/cobra` for command-line interface
  - `github.com/ZanzyTHEbar/errbuilder-go` for custom error handling
  - `github.com/ZanzyTHEbar/assert-lib` for assertions
  - `github.com/spf13/viper` for configuration management
  - `github.com/google/uuid` for UUID generation
  - `github.com/stretchr/testify` for testing utilities

ALL PLACEHOLDER LOGIC MUST BE IMPLEMENTED IN REAL-TIME. DO NOT LEAVE ANY PLACEHOLDER LOGIC UNIMPLEMENTED. DO NOT LEAVE ANY TODOs UNADDRESSED. PRIORITISE COMPLETING ALL TASKS IN REAL-TIME BEFORE MOVING ON TO NEW TASKS.

- make sure to use the following packages for testing:
  - `github.com/stretchr/testify/assert` for assertions in tests
  - `github.com/stretchr/testify/require` for requirements in tests
  - `github.com/stretchr/testify/suite` for test suites
- make sure to use the following packages for mocking:
  - `github.com/stretchr/testify/mock` for mocking in tests
  - `github.com/golang/mock/gomock` for generating mocks
- make sure to use the following packages for logging:
  - `github.com/sirupsen/logrus` for structured logging
  - `github.com/rs/zerolog` for zero-allocation logging
- make sure to use the following packages for configuration:
  - `github.com/spf13/viper` for configuration management
  - `github.com/joho/godotenv` for loading environment variables from `.env` files
- make sure to use the following packages for database interactions:
  - `github.com/jmoiron/sqlx` for SQL database interactions
  - `github.com/lib/pq` for PostgreSQL driver
  - `github.com/go-sql-driver/mysql` for MySQL driver
  - `github.com/mattn/go-sqlite3` for SQLite driver
- make sure to use the following packages for HTTP interactions:
