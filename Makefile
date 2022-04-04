# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
EXENAME := sc
BIN_NAME=./bin/sc
BIN_NAME=${EXEDIR}/${EXENAME}

# This .PHONY target will ignore any file that exists with the same name as the target
# in your makefile, and build it regardless.
.PHONY: all init genstubs build run clean

# The all target is the default target when make is called without any arguments.
# This is because it is the first target whose name does not start with a '.'
all: clean | run

init:
	- rm go.mod
	- rm go.sum
	go mod init github.com/samirgadkari/sidecar
	go mod tidy -compat=1.17

genstubs:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       protos/v1/messages/sidecar.proto

${EXEDIR}:
	mkdir ${EXEDIR}

build: | ${EXEDIR}
	go build -o ${BIN_NAME} pkg/main/main.go

run: build
	./${BIN_NAME} serve

test:
	alacritty --working-directory ~/work/do/search/sidecar -e ./${BIN_NAME} serve &
	go test -v ./...
	killall ${EXENAME}

clean:
	go clean
	rm ${BIN_NAME}
	go clean -cache -modcache -i -r
	go mod tidy -compat=1.17
