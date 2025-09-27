FROM golang:1.25.1-alpine

RUN apk add --no-cache vips-dev pkgconf 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o main .

CMD ["./main"]
