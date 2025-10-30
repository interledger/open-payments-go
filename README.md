# Development

**Requires:**

- go1.21+

## OpenAPI Specs

This repository contains a Git submodule, which contains the Open Payments OpenAPI specifications.
After cloning, make sure to initialize and update it:

```bash
git submodule update --init
```

Alternatively, clone the repository with submodules in one step:

```bash
git clone --recurse-submodules https://github.com/interledger/open-payments-go.git
```

The SDK depends on types generated from the specs. To generate types from the specs:

    go generate ./generated

## Commands

### Run tests

From root:

    go test ./...

Note, this runs all tests including integration, which requires the Rafiki localenv.

Integration tests can be run against a `local` or `testnet` Rafiki environment:

    go test ./test/integration -env=testnet

Or to run test from a specific package:

    go test ./httpsignatureutils

To include all logs for debugging or development:

    go test -v ./...
