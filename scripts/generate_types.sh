#!/bin/bash

echo "Generating schemas..."
oapi-codegen -generate types,skip-prune /Users/blaircurrey/code/interledger/open-payments-go/api/schemas.yaml > generated/schemas/types.go

echo "Generating resource server types..."
oapi-codegen -package resourceserver -generate types --import-mapping=./schemas.yaml:github.com/interledger/open-payments-go/pkg/generated/schemas /Users/blaircurrey/code/interledger/open-payments-go/api/resource-server-edited.yaml > generated/resourceserver/types.go

echo "Generating auth server types..."
oapi-codegen -package authserver -generate types --import-mapping=./schemas.yaml:github.com/interledger/open-payments-go/pkg/generated/schemas /Users/blaircurrey/code/interledger/open-payments-go/api/auth-server.yaml > generated/authserver/types.go

echo "Generating wallet server types..."
oapi-codegen -package walletaddressserver -generate types --import-mapping=./schemas.yaml:github.com/interledger/open-payments-go/pkg/generated/schemas /Users/blaircurrey/code/interledger/open-payments-go/api/wallet-address-server-edited.yaml > generated/walletaddressserver/types.go

echo "Type generation complete."
