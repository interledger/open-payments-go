# Development

requires:

- go1.21+

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

### Update OpenAPI Spec

When you want to pull in new open payments openapi specs, use `fetch_spec.sh` script to get the `open-payments` Open API specs by [release tag](https://github.com/interledger/open-payments/releases). For example:

    ./scripts/fetch_spec.sh @interledger/open-payments@5.0.0

### Generate OpenAPI types

After pulling in new open payments openapi specs, use `generate_types.sh` to generate all the open api types from the spec files. Type generation is handled by `deepmap/oapi-codegen` and outputs to the `generated` directory.

Run:

    ./scripts/generate_types.sh

> [!NOTE]
> Each spec has is it's own package for the generated types because of name collisions between some spec files (multiple declarations of `Incoming Payment`, `GNAPScopes`, etc.).

> [!NOTE]
> Currently there are some quirks with our schemas and how the types are generated which require a few manual fixes before generating types from the fetched specs. This is reflected by the `api/*-edited/yaml` files and the `./scripts/generate_types script`. These manual fixes are:
>
> - `resource-server.yaml`: remove refs and inline `optional-signature` and `optional-signature-input` in `signature` and `signature-input`.
> - `wallet-address-server`: remove `additionalProperties: true` `from WalletAddress`
>   - `additionalProperties` are true by default so behavior doesnt change. More details on on how `deepmap/oapi-codegen` handles this field: https://github.com/deepmap/oapi-codegen?tab=readme-ov-file#additional-properties-in-type-definitions
