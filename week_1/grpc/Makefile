LOCAL_BIN:=$(CURDIR)/bin/

.PHONY: help
help: ## Показать справку по доступным командам
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<цель>\033[0m\n\nДоступные цели:\n"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

install-deps: ## Установить зависимости, необходимые для генерации кода из proto-файлов
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

get-deps: ## Обновить зависимости proto-компилятора и gRPC
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate: ## Сгенерировать код для всех API, связанных с заметками, с использованием плагинов protoc-gen-go и protoc-gen-go-grpc
	make generate-note-api

generate-note-api: ## Сгенерировать Go-код для API note_v1 из proto-файла с использованием плагинов protoc-gen-go и protoc-gen-go-grpc
	mkdir -p pkg/note_v1
	protoc --proto_path api/note_v1 \
	--go_out=pkg/note_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/note_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/note_v1/note.proto
