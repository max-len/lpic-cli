
FROM golang:alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN pwd
COPY cmd cmd
COPY internal internal
RUN mkdir -p /client
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /client/client ./cmd/client
ENTRYPOINT ["/client/client"]
