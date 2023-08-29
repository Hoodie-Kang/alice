FROM golang:1.19.4

WORKDIR /go/src/alice

COPY . .

RUN cd example && go build -o p2p_test ./sign.go

CMD ["./example/p2p_test"]