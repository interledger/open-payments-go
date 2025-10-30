package generate

//go:generate go tool oapi-codegen --config authserver.config.yaml ../open-payments-specifications/openapi/auth-server.yaml
//go:generate go tool oapi-codegen --config walletaddressserver.config.yaml ../open-payments-specifications/openapi/wallet-address-server.yaml
//go:generate go tool oapi-codegen --config resourceserver.config.yaml ../open-payments-specifications/openapi/resource-server.yaml
