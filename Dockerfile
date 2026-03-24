FROM golang:1.24-trixie@sha256:5835f052b784aa39f2fe9070def3568605c8bc3fcd810f10402066348b61e716 as builder

ADD . /webtunnel

ENV CGO_ENABLED=0

WORKDIR /webtunnel

RUN go build -ldflags="-s -w" -o "build/server" gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/webtunnel/main/server

FROM containers.torproject.org/tpo/tpa/base-images/debian:trixie@sha256:e5b2f3924d7247e6e2a04f1a13b3c8b617fd83a0d7c1f4c9e5b4f3de7e8e48dd

COPY --from=builder /webtunnel/build/server /usr/bin/webtunnel-server

# Install dependencies to add Tor's repository.
RUN apt-get update && apt-get install -y \
    wget \
    gpg \
    gpg-agent \
    ca-certificates \
    libcap2-bin \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# See: <https://2019.www.torproject.org/docs/debian.html.en>
RUN wget -qO- https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --dearmor | tee /usr/share/keyrings/deb.torproject.org-keyring.gpg >/dev/null
RUN groupadd -g 101 debian-tor
RUN useradd --system -u 101 -g 101 -s /usr/bin/nologin -d /var/lib/tor debian-tor

RUN echo "deb [signed-by=/usr/share/keyrings/deb.torproject.org-keyring.gpg] https://deb.torproject.org/torproject.org trixie main\n" >> /etc/apt/sources.list.d/tor.list

# Install remaining dependencies.
RUN apt-get update && apt-get install -y \
    tor \
    tor-geoipdb \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Our torrc is generated at run-time by the script start-tor.sh.
RUN rm /etc/tor/torrc
RUN chown debian-tor:debian-tor /etc/tor
RUN chown debian-tor:debian-tor /var/log/tor

ADD release/container/start-tor.sh /usr/local/bin
RUN chmod 0755 /usr/local/bin/start-tor.sh

ADD release/container/get-bridge-line.sh /usr/local/bin
RUN chmod 0755 /usr/local/bin/get-bridge-line.sh

ENTRYPOINT ["/usr/local/bin/start-tor.sh"]
