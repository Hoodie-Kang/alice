# Build stage
FROM golang:1.19.4 AS build

WORKDIR /go/src

COPY . .

RUN cd example && go build -o test ./sign.go

# Run stage
FROM golang:1.19.4 AS run

WORKDIR /go/src

COPY --from=build /go/src/example/test /go/src/

CMD ["./test"]