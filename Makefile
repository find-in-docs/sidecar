# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
BIN_NAME=./bin/sc

# This .PHONY target will ignore any file that exists with the same name as the target
# in your makefile, and build it regardless.
.PHONY: all init genstubs build run clean

# The all target is the default target when make is called without any arguments.
all: clean | run

init:
	go mod init github.com/samirgadkari/sidecar
	mkdir cli
	cd cli && cobra init
	cd cli && cobra add serve

genstubs:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       protos/v1/messages/sidecar.proto

${EXEDIR}:
	mkdir ${EXEDIR}

build: | ${EXEDIR}
	go build -o ${BIN_NAME} cli/main.go

run: build
	./${BIN_NAME} serve

clean:
	go clean
	rm ${BIN_NAME}
	go clean -cache -modcache -i -r
	go mod tidy
