#!/bin/bash

pwd
ls
cd gateway && rm go.sum && rm go.mod
cd run && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway . && \
    chmod +x gateway && docker build -t gateway:v1.0 .
cd ../../

cd item && rm go.sum && rm go.mod
cd run && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o item . && \
    chmod +x item && docker build -t item:v1.0 .
cd ../../

cd order && rm go.sum && rm go.mod
cd run && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order . && \
    chmod +x order && docker build -t order:v1.0 .
cd ../../

cd user && rm go.sum && rm go.mod
cd run && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user . && \
    chmod +x user && docker build -t user:v1.0 .
cd ../../

cd view && docker build -t view:v1.0 .
cd -

ls
