FROM caddy:2-builder AS builder

WORKDIR /srv

COPY . .

RUN env CGO_ENABLED=0 xcaddy build \
    --with github.com/HeavenVolkoff/caddy-cloudflare-only=$PWD \
    --output /usr/bin/caddy

FROM caddy:2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
