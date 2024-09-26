# Caddy + Cloudflare

Caddy v2 module for retrieving Cloudflare IP blocks. It extends [trusted_proxy](https://caddyserver.com/docs/caddyfile/options#trusted-proxies) with a new `cloudflare` IP source module and add a new `cloudflare_only` directive for rejecting requests from ips outside the Cloudflare block range.

## Installation

To install the module, you need to build Caddy with it included. You can do this using `xcaddy`:

```sh
xcaddy build --with github.com/HeavenVolkoff/caddy-cloudflare
```

## Configuration

Add the following to your Caddyfile to enable the module:

```caddyfile
{
    order cloudflare_only before redir
    servers {
        trusted_proxies cloudflare
    }
}

yourdomain.com {
    cloudflare_only
    respond "Hello, World!"
}
```

### `cloudflare_only` Options

- `reject_if_empty`: (boolean) If set to `false`, the module will allow every request while the IP blocks are not yet populated.

## Usage

Once configured, the module will automatically fetch the latest Cloudflare IP blocks, add it as a trusted upstream proxy, and restrict communication to only remote ip know to be from Cloudflare. If a request comes from an IP not in the Cloudflare block range, the connection will be rejected with a 403 - Forbidden.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with your improvements.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Copyright

I am not affiliated with either Caddy or Cloudflare.

Caddy® is a registered trademark of Stack Holdings GmbH.
Cloudflare® is a registered trademarks of Cloudflare, Inc.
