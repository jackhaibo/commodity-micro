FROM ubuntu:16.04
LABEL MAINTAINER jackhaibo

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user -v main.go && \
    chmod +x user
   
COPY user /usr/bin