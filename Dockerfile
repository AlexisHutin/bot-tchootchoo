FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o tchoo-tchoo .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY --from=build /app/tchoo-tchoo .

RUN ldd /app/tchoo-tchoo

RUN chmod +x /app/tchoo-tchoo

CMD ["sh"]