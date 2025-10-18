FROM golang:1.25 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Final stage
FROM public.ecr.aws/lambda/provided:al2023

# Copy the binary from builder
COPY --from=builder /app/main /main

# Set the entrypoint
ENTRYPOINT ["/main"]
