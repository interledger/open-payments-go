# Integration Tests

This package contains integration tests for the Open Payments client against a running instance of [Rafiki](https://github.com/interledger/rafiki). The tests can be run against the Rafiki local environment or [testnet](https://wallet.interledger-test.dev/).

# Testnet Environment Configuration

Environment variables for the sending testnet wallet address must be set in a `.env.testnet` file before running the tests. See `.env.testnet.example`. This is required for initializing the authenticated client and completing the interaction step for the outgoing payment grant.

> [!NOTE]
> The sending wallet address needs to be manually funded on testnet, otherwise the incoming payment will never complete.

## Running the Tests

To run all integration tests against testnet:

```bash
go test ./test/integration -v -env=testnet
```

Or a local rafiki environemnt:

```bash
go test ./test/integration -v -env=local
```
