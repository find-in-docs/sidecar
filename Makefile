# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
BIN_NAME=./bin/sc

# The all target is the default target when make is called without any arguments.
all: run

${EXEDIR}:
	mkdir ${EXEDIR}

build: | ${EXEDIR}
	go build -o ${BIN_NAME} cli/main.go

run: build
	./${BIN_NAME} server

clean:
	go clean
	rm ${BIN_NAME}
	go clean -cache -modcache -i -r
	go mod tidy
