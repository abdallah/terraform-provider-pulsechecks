BINARY=terraform-provider-pulsechecks
INSTALL_DIR=~/.terraform.d/plugins/registry.terraform.io/abdallah/pulsechecks/0.1.0/linux_amd64

.PHONY: build test install fmt

build:
	go build -o $(BINARY) .

test:
	go test ./internal/provider/... -v -timeout 120s

testacc:
	TF_ACC=1 go test ./internal/provider/... -v -timeout 120s

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

fmt:
	gofmt -w .
