FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /farelah-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /farelah-api /usr/local/bin/farelah-api
EXPOSE 8080
ENTRYPOINT ["farelah-api"]

