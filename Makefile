include .env
GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=D:\Program Files\Git\bin\bash.exe
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell "$(Git_Bash)" -c "find api -name *.proto")
	ENT_DIRS=$(shell $(Git_Bash) -c "find ./app -type d -name ent | sed 's|^\./||'")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
	ENT_DIRS=$(shell find ./app -type d -name ent | sed 's|^\./||')
endif

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install entgo.io/ent/cmd/ent@v0.14.4
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

.PHONY: api
OPENAPI_OUT_DIR=./docs/openapi
# generate api proto
api:
	protoc --proto_path=./api \
		   --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
		   --go-http_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:./docs \
	       $(API_PROTO_FILES)

.PHONY: ent
ent:
ifeq ($(GOHOSTOS), windows)
	@for %%i in ($(ENT_DIRS)) do \
		go generate ./%%i
else
	@for d in $(ENT_DIRS); do \
		go generate ./$$d
	done
endif 

.PHONY: migrations
migrations:
ifeq ($(GOHOSTOS), windows)
	@for %%i in ($(ENT_DIRS)) do \
		cd %%i && cd .. && \
		atlas migrate diff $(tag) --to "ent://ent/schema" \
		--dir "file://ent/migrate/migrations" \
		--dev-url "$(DEV_DB_ADDR)"
else
	@for d in $(ENT_DIRS); do \
		cd $$d && cd .. && \
		atlas migrate diff $(tag) --to "ent://$$d/schema" \
		--dir "file://$$d/migrate/migrations" \
		--dev-url "$(DEV_DB_ADDR)"
	done
endif

.PHONY: hash-migrations
hash-migrations:
ifeq ($(GOHOSTOS), windows)
	@for %%i in ($(ENT_DIRS)) do \
		cd %%i && cd .. && \
		atlas migrate hash \
		--dir "file://ent/migrate/migrations"
else
	@for d in $(ENT_DIRS); do \
		cd $$d && cd .. && \
		atlas migrate hash $(tag) \
		--dir "file://$$d/migrate/migrations"
	done
endif

.PHONY: migrate
migrate:
ifeq ($(GOHOSTOS), windows)
	@for %%i in ($(ENT_DIRS)) do \
		cd %%i && cd .. && \
		atlas migrate apply \
		--dir "file://ent/migrate/migrations" \
		--url "$(DB_ADDR)"
else
	@for d in $(ENT_DIRS); do \
		cd $$d && cd .. && \
		atlas migrate apply \
		--dir "file://ent/migrate/migrations" \
		--url "$(DB_ADDR)"
	done
endif

