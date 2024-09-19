package cloudflare

import (
	"net"
	"net/http"
	"net/netip"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

type CloudflareOnly struct {
	RejectIfEmpty bool `json:"reject_if_empty,omitempty"`

	block  *CloudflareIPBlock
	logger *zap.Logger
}

func (CloudflareOnly) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.cloudflare_only",
		New: func() caddy.Module { return new(CloudflareOnly) },
	}
}

func (cf *CloudflareOnly) Provision(ctx caddy.Context) error {
	cf.block = GetCloudflareIpBlock(ctx)
	cf.logger = ctx.Logger()

	return nil
}

func (cf *CloudflareOnly) Validate() error {
	return nil
}

func (cf CloudflareOnly) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	cf.block.lock.RLock()
	defer cf.block.lock.RUnlock()

	if len(cf.block.ips) > 0 {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return err
		}

		ip, err := netip.ParseAddr(remoteIP)
		if err != nil {
			return err
		}

		for _, network := range cf.block.ips {
			if network.Contains(ip) {
				return next.ServeHTTP(w, r)
			}
		}

		cf.logger.Debug("rejected request", zap.String("remote_ip", remoteIP), zap.String("url", r.URL.String()))
	} else if cf.RejectIfEmpty {
		cf.logger.Debug("Cloudflare IP list is empty, rejecting all request")
	} else {
		return next.ServeHTTP(w, r)
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return nil
}

func (cf *CloudflareOnly) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	cf.RejectIfEmpty = true
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "reject_if_empty":
				var arg string
				if !d.Args(&arg) {
					return d.ArgErr()
				}

				if arg == "false" {
					cf.RejectIfEmpty = false
				} else if arg != "true" {
					return d.Errf("invalid argument: %s", arg)
				}
			}
		}
	}

	return nil
}

var (
	_ caddy.Validator             = (*CloudflareOnly)(nil)
	_ caddy.Provisioner           = (*CloudflareOnly)(nil)
	_ caddyfile.Unmarshaler       = (*CloudflareOnly)(nil)
	_ caddyhttp.MiddlewareHandler = (*CloudflareOnly)(nil)
)
