FROM golang:1.20-alpine
WORKDIR /src
COPY . .
WORKDIR /src/cmd/
RUN CGO_ENABLED=0 go build -o dnscheck

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src/cmd/dnscheck /opt/dnscheck/dnscheck
ENTRYPOINT ["/opt/dnscheck/dnscheck"]