FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /printsec-api ./main.go

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser && mkdir -p /app/uploads
COPY --from=build /printsec-api /app/printsec-api
COPY migrations /app/migrations
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
CMD ["/app/printsec-api"]

