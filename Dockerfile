FROM golang as builder

WORKDIR /rnelson3-agent
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /rnelson3-agent
COPY --from=builder /rnelson3-agent/app .

CMD ["./app"]