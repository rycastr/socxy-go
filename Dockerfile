FROM golang:1.16.3-alpine3.13

WORKDIR /go/src/socxy-server

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["socxy-server"]