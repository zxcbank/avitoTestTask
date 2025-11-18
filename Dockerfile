FROM golang:1.25

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main ./cmd/avitoTestTask

RUN chmod +x /app/main

COPY internal/storage/migrations /app/migrations

EXPOSE 8080
CMD ["./main"]