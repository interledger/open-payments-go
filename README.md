# Development

requires:

- go1.21+

## Commands

### Run tests

From root:

    go test ./...

Note, this runs all tests including integration, which requires the Rafiki localenv.

To just run the integration tests:

    go test ./test/integration

Or to run test from a specific package:

    go test ./httpsignatureutils

To include all logs for debugging or development:

    go test -v ./...

### OpenAPI Specs

This repository uses [open-payments-specifications](https://github.com/interledger/open-payments-specifications) as a submodule.

Generate types from the specs:

    go generate ./generated
