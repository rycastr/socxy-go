FROM golang:1.16.3-alpine3.13

WORKDIR /go/src/socxy-server

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN apk add cmake clang make g++ git

RUN git clone https://github.com/ambrop72/badvpn.git

RUN cd badvpn && mkdir build && cd build && \
    cmake .. -DBUILD_NOTHING_BY_DEFAULT=1 -DBUILD_UDPGW=1 && \
    make install

CMD ["socxy-server"]