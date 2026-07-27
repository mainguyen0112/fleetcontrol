SPEC := openapi/openapi.yaml
GEN_DIR := gen
PACKAGE := gen

.PHONY: generate generate-models generate-server generate-client verify-generate

generate-models:
	go tool oapi-codegen -generate models -package $(PACKAGE) -o $(GEN_DIR)/models.gen.go $(SPEC)

generate-server:
	go tool oapi-codegen -generate chi-server -package $(PACKAGE) -o $(GEN_DIR)/server.gen.go $(SPEC)

generate-client:
	go tool oapi-codegen -generate client -package $(PACKAGE) -o $(GEN_DIR)/client.gen.go $(SPEC)

generate: generate-models generate-server generate-client

verify-generate: generate
	@git diff --exit-code -- $(GEN_DIR) || \
		(echo "Generated code khong khop voi openapi.yaml. Hay chay make generate va commit lai." && exit 1)
