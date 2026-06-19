FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/presupuesto-api ./cmd/api

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
WORKDIR /app
COPY --from=build /bin/presupuesto-api /app/presupuesto-api
EXPOSE 8080
CMD ["/app/presupuesto-api"]
