FROM ubuntu:16.04
LABEL MAINTAINER jackhaibo

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o item -v main.go && \
    chmod +x item
   
COPY item /usr/bin