# Development

requires:

- go1.20

## Commands

### Run all tests

    go test ./...

### Update OpenAPI Spec

Use `fetch_spec.sh` script to get the `open-payments` Open API specs by [release tag](https://github.com/interledger/open-payments/releases):

    ./scripts/fetch_spec.sh @interledger/open-payments@5.0.0
