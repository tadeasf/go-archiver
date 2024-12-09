FROM golang:1.21

WORKDIR /app

COPY . .

RUN go mod tidy

RUN gopy build -output=./_binding -vm=python3 .
