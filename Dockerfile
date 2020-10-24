FROM golang:1.15 as builder
WORKDIR /app

COPY . /app
RUN adduser --disabled-password --gecos '' slim && \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-w -s" -o qrator-metrics

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/qrator-metrics /bin/qrator-metrics

EXPOSE 9805
USER slim
WORKDIR /

ENTRYPOINT ["/bin/qrator-metrics"]