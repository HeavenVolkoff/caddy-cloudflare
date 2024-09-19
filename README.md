# Caddy Cloudflare Only Plugin

Caddy v2 module for rejecting requests outside Cloudflare IP blocks.

## Installation

To install the plugin, you need to build Caddy with the plugin included. You can do this using `xcaddy`:

```sh
xcaddy build --with github.com/HeavenVolkoff/caddy-cloudflare-only
```

## Configuration

Add the following to your Caddyfile to enable the plugin:

```caddyfile
{
    order cloudflare_only before redir
}

yourdomain.com {
    cloudflare_only
    respond "Hello, World!"
}
```

### Options

- `reject_if_empty`: (boolean) If set to `false`, the plugin will allow every request while the IP blocks are not yet populated.

## Usage

Once configured, the plugin will automatically fetch the latest Cloudflare IP blocks and enforce the IP restrictions. If a request comes from an IP not in the Cloudflare IP blocks, the connection will be rejected with a 403 - Forbidden.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with your improvements.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
