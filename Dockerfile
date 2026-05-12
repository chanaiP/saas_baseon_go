FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/saas-api ./cmd/api

FROM scratch

COPY --from=build /out/saas-api /app/saas-api
COPY --from=build /src/internal/apps /app/internal/apps
EXPOSE 8081
CMD ["/app/saas-api"]
