FROM golang:1.19.4

WORKDIR /go/src/alice

COPY . .

RUN cd example/cggmp/dkg && go build -o dkg_test ./main.go

CMD ["./example/cggmp/dkg/dkg_test"]