FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cbt-api ./main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /app/cbt-api /usr/local/bin/cbt-api
EXPOSE 8080
CMD ["cbt-api"]
