# Build stage
FROM golang:1.19.4 AS build

WORKDIR /go/src

COPY . .

RUN go build ./wait.go && cd example && go build -o sign ./sign.go

# Run stage
FROM golang:1.19.4 AS run

WORKDIR /go/src

COPY --from=build /go/src/example/sign /go/src/wait /go/src/

CMD ["./wait"]