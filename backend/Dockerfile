FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /bin/server /usr/local/bin/server
USER nobody
EXPOSE 8080
CMD ["server"]
