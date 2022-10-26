FROM golang as builder

WORKDIR /reddit-worker
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /reddit-worker
COPY --from=builder /reddit-worker/app .
COPY --from=builder /reddit-worker/.env .

CMD ["./app"]