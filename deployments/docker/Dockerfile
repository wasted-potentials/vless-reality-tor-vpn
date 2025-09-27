# --- builder ---
FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/vpnd ./cmd/vpnd && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/keygen ./cmd/keygen

# --- runtime ---
FROM debian:bookworm-slim
ARG XRAY_VERSION=v1.8.7
ENV XRAY_ZIP=Xray-linux-64.zip
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl unzip tzdata tor libcap2-bin && \
    rm -rf /var/lib/apt/lists/* && \
    curl -fsSL -o /tmp/${XRAY_ZIP} https://github.com/XTLS/Xray-core/releases/download/${XRAY_VERSION}/${XRAY_ZIP} && \
    unzip -d /usr/local/bin /tmp/${XRAY_ZIP} xray && rm /tmp/${XRAY_ZIP} && \
    setcap 'cap_net_bind_service=+ep' /usr/local/bin/xray

# create user
RUN useradd -ms /bin/bash vpn && mkdir -p /var/lib/vpnd /etc/vpnd && chown -R vpn:vpn /var/lib/vpnd /etc/vpnd
COPY --from=builder /out/vpnd /usr/local/bin/vpnd
COPY --from=builder /out/keygen /usr/local/bin/keygen
COPY example/config.yaml /etc/vpnd/config.yaml
COPY deployments/tor/torrc /etc/tor/torrc
USER vpn
EXPOSE 443 8080 9100
ENTRYPOINT ["/usr/local/bin/vpnd","-config","/etc/vpnd/config.yaml"]
