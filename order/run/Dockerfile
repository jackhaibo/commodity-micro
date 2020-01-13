FROM ubuntu:16.04
LABEL MAINTAINER jackhaibo

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order -v main.go && \
    chmod +x order
   
COPY order /usr/bin