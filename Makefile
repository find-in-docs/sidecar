# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
EXENAME := sc
BIN_NAME=./${EXEDIR}/${EXENAME}

# The .PHONY target will ignore any file that exists with the same name as the target
# in your makefile, and built it regardless.
.PHONY: all init genstubs build run clean upload

# The all target is the default target when make is called without any arguments.
all: clean | run

init:
	echo "Setting up local ..."
	go env -w GOPROXY=direct 
	echo "To get protoc, look here:"
	echo "  example: https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-x86_64.zip"
	echo "To install protoc-gen-go-grpc, do this:"
	echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	echo "To install protoc-gen-go, do this:"
	echo "  go install google.golang.org/grpc/cmd/protoc-gen-go"
	- rm go.mod
	- rm go.sum
	go mod init github.com/find-in-docs/sidecar
	go mod tidy

genstubs: init
	echo "Generating local stubs ..."
	protoc --go_out=. --go_opt=paths=source_relative \
   			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	 			protos/v1/messages/sidecar.proto

${EXEDIR}:
	echo "Building exe directory ..."
	mkdir ${EXEDIR}

build: genstubs | ${EXEDIR}
	echo "Building locally ..."
	go build -o ${BIN_NAME} pkg/main/main.go

run: build
	echo "Running locally ..."
	./${BIN_NAME} serve

test:
	alacritty --working-directory ~/work/do/search/sidecar -e ./${BIN_NAME} serve &
	go test -v ./...
	killall ${EXENAME}

clean:
	echo "Cleaning locally ..."
	go clean
	- rm ${BIN_NAME}
	go clean -cache -modcache -i -r
	go mod tidy

upload: init
	echo "Start building on minikube ..."
	docker build -t sidecar -f ./Dockerfile .
	
	# go env -w GOPROXY="https://proxy.golang.org,direct"
	# echo "Get each of these packages in the Dockerfile"
	# rg --iglob "*.go" -o -I -N "[\"]github([^\"]+)[\"]" | sed '/^$/d' | sed 's/\"//g' | awk '{print "RUN go get " $0}'
	
	# We specify image-pull-policy-Never because we're actually building the image on minikube.
	# kubectl run sidecar --image=sidecar:latest --image-pull-policy=Never --restart=Never
	
	# kubectl apply -f manifests/minikube
