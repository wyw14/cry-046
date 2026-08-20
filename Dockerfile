FROM golang:1.24-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./cmd/server

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=backend /out/server /app/server
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
