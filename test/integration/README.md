# Integration Tests

This package contains integration tests for the Open Payments client against a running instance of [Rafiki](https://github.com/interledger/rafiki).

Currently, the tests are configured to run against **Rafiki's local Docker environment**. Support for running against **testnet** is planned in the future.

## Running the Tests

To run all integration tests:

```bash
go test ./test/integration -v
```
