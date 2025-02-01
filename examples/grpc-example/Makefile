PROTO_FILES := $(wildcard *.proto)
GO_OUT_DIR := .

.PHONY: generate
generate:
	protoc --go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

.PHONY: clean
clean:
	rm -f calculator/*.go 