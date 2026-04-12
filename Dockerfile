# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-gateway ./cmd/api-gateway


FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 65532 nonroot

WORKDIR /app

COPY --from=builder /out/api-gateway .

USER nonroot:nonroot

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["./api-gateway"]
