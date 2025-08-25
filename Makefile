
.PHONY: gen-proto

buf-gen-proto:
	buf generate --exclude-path "./vendor.protogen/"

gen-ugc-api-v1:
	mkdir -p pkg/ugcservice/v1
	protoc --proto-path api/ugcservice/v1 --proto-path vendor.protogen \
	--go-out=pkg/ugcservice/v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/ugcservice/v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-grpc=bin/protoc-gen-go-grpc \
	--validate_out lang=go:pkg:ugcservice/v1 --validate_opt=paths=source_relative \
	--plugin=protoc-gen-validate=bin/protoc-gen-validate \
	--grpc-gateway_out=pkg/ugcservice/v1 --grpc-gateway_opt=paths=source_relative \
	--plugin=protoc-gen-grpc-gateway=bin/protoc-gen-grpc-gateway \
	api/ugcservice/v1/ugc.proto


vendor-proto:
	@if [ ! -d vendor.protogen/validate ]; then \
		mkdir -p vendor.protogen/validate &&\
		git cline https://github.com/envoyproxy/protoc-gen-validate vendor.protogen/protoc-gen-validate \
		mv vendor.protogen/protoc-gen-validate/validate/*.proto vendor.protogen/validate &&\
		rm -rf vendor.protogen/protoc-gen-validate ;\
	fi
	@if [ ! -d vendor.protogen/google ]; then \
		git clone https://github.com/googleapis/googleapis vendor.protogen/googleapis &&\
		mkdir -p vendor.protogen/google/ &&\
		mv vendor.protogen/googleapis/google/api/ vendor.protogen/google &&\
		rm -rf vendor.protogen/googleapis ;\
	fi