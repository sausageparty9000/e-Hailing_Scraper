FROM golang:1.24-alpine AS source
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

FROM source AS test
CMD ["go", "test", "./...", "-v", "-cover", "-count=1"]

FROM source AS build
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /farelah-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /farelah-api /usr/local/bin/farelah-api
EXPOSE 8080
ENTRYPOINT ["farelah-api"]
