package generate

//go:generate go tool oapi-codegen --config authserver.config.yaml ../open-payments-specifications/openapi/auth-server.yaml
//go:generate go tool oapi-codegen --config walletaddresserver.config.yaml ../open-payments-specifications/openapi/wallet-address-server.yaml
