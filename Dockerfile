FROM golang:1.25.1-alpine

RUN apk add --no-cache vips 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]
