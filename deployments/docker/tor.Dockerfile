FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install -y --no-install-recommends tor tor-geoipdb ca-certificates \
    && rm -rf /var/lib/apt/lists/*
# если хочешь фиксировать конфиг в образе:
# COPY deployments/tor/torrc /etc/tor/torrc
USER debian-tor
EXPOSE 9050
ENTRYPOINT ["tor","-f","/etc/tor/torrc"]
