package cloudflare

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(CloudflareIpSource{})
	caddy.RegisterModule(CloudflareOnly{})
	httpcaddyfile.RegisterHandlerDirective("cloudflare_only", parseCaddyfileCloudflareOnly)
}

func parseCaddyfileCloudflareOnly(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var cf CloudflareOnly
	err := cf.UnmarshalCaddyfile(h.Dispenser)
	return cf, err
}
