package cloudflare

import (
	"net/http"
	"net/netip"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

type CloudflareIpSource struct {
	block  *CloudflareIPBlock
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (CloudflareIpSource) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.ip_sources.cloudflare",
		New: func() caddy.Module { return new(CloudflareIpSource) },
	}
}

func (cf *CloudflareIpSource) Provision(ctx caddy.Context) error {
	cf.block = GetCloudflareIpBlock(ctx)
	cf.logger = ctx.Logger()

	return nil
}

func (cf *CloudflareIpSource) GetIPRanges(_ *http.Request) []netip.Prefix {
	cf.block.lock.RLock()
	defer cf.block.lock.RUnlock()
	return cf.block.ips
}

func (cf *CloudflareIpSource) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // Skip module name.
	return nil
}

var (
	_ caddy.Module            = (*CloudflareIpSource)(nil)
	_ caddy.Provisioner       = (*CloudflareIpSource)(nil)
	_ caddyfile.Unmarshaler   = (*CloudflareIpSource)(nil)
	_ caddyhttp.IPRangeSource = (*CloudflareIpSource)(nil)
)
