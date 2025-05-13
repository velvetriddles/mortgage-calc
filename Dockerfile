FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o mortgage-calc ./cmd/app

FROM scratch

COPY --from=builder /app/mortgage-calc /app/mortgage-calc
COPY --from=builder /app/config.yml /app/config.yml

WORKDIR /app

CMD ["/app/mortgage-calc"] 