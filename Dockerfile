
# Base image
FROM golang:1.19.5-alpine3.17

# Specify work directory on the image.
# All commands will refer to this work directory from now on below.
WORKDIR /app
# ENV GOROOT=/app

# Copy local go.mod and go.sum files into the image
COPY go.mod ./
COPY go.sum ./

# Clean the modcache
# This is not required all the time. You should run this only
# when your modcache contains older versions that you cannot upgrade for some reason.
# RUN go clean -cache -modcache -i -r

# This should always be set. This way, instead of getting packages from Google servers,
# you get them directly from github. This way, there is no sync lag between
# your changes on github and your packages on Google servers. Sometimes, it takes
# more than a day for Google servers to catch up to your changes.
# This is now updated in the Makefile under init.
# RUN go env -w GOPROXY=direct 

RUN apk update && \
    apk add git && \
    apk add tree && \
    apk add protoc && \
    apk add protobuf-dev && \
    apk add --update util-linux

# Copy package source code
RUN mkdir -p protos/v1/messages
COPY pkg/ /app/pkg/
COPY protos/ /app/protos/

# This file contains the DNS server information. It is used by the persistlogs
# service to:
#   - Complete the Fully Qualified Domain Name of the request
#   - Locate the IP address of the DNS server
COPY manifests/minikube/resolv.conf /etc/resolv.conf

# Git config change to get over the issue of "Could not resolve github.com".
# This was added to .bashrc instead.
# RUN git config --global --unset http.proxy 
# RUN git config --global --unset https.proxy

# Download required packages in the image
RUN go mod download
RUN go mod tidy

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# COPY $GOPATH/include/ /go/
# COPY home/samirgadkari/go/include /go/
# RUN ls -l /go

# Build stubs inside minikube
RUN protoc \
        --go_out=. --go_opt=paths=source_relative \
   			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	 			protos/v1/messages/sidecar.proto

# RUN tree /app
# RUN ls -l /app
RUN go build -o sidecar ./pkg/main/main.go

CMD ["./sidecar"]
