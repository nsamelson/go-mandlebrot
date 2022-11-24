FROM golang:1.19.3 as base

FROM base as dev

# RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
RUN go install github.com/cosmtrek/air@latest

WORKDIR /opt/app/api
CMD "air"