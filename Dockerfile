FROM golang:latest

WORKDIR reddit-worker

COPY . .

CMD ["go", "run", "main.go"]