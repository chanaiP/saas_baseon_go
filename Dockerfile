FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/saas-api ./cmd/api

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/saas-api /app/saas-api
EXPOSE 8081
CMD ["/app/saas-api"]
