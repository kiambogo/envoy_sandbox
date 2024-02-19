# Build stage
FROM golang:1.22 AS builder

RUN apt update
RUN apt install -y protobuf-compiler

WORKDIR /app

COPY proto/hello.proto ./proto/hello.proto
COPY main.go .
COPY server.go .
COPY client.go .
COPY go.mod .
COPY go.sum .

RUN go get -u google.golang.org/grpc
RUN go install github.com/golang/protobuf/protoc-gen-go@v1.5.3
RUN protoc --go_out=plugins=grpc:proto ./proto/hello.proto

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=builder /app/app /app

CMD ["/app"]
