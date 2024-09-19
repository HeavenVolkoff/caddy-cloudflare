package cloudflare_only

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

type CloudflareIPs struct {
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
}

func fetchCloudflareIPs(ctx context.Context) (*CloudflareIPs, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.cloudflare.com/client/v4/ips", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result CloudflareIPs `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Result, nil
}

func init() {
	caddy.RegisterModule(CloudflareOnly{})
	httpcaddyfile.RegisterHandlerDirective("cloudflare_only", parseCaddyfile)
}

type CloudflareOnly struct {
	RejectIfEmpty bool `json:"reject_if_empty,omitempty"`

	mu     *sync.RWMutex
	logger *zap.Logger

	IPBlocks []netip.Prefix
}

func (CloudflareOnly) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.cloudflare_only",
		New: func() caddy.Module { return new(CloudflareOnly) },
	}
}

func (cf *CloudflareOnly) updateIPBlocks(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ips, err := fetchCloudflareIPs(ctx)
			if err != nil {
				cf.logger.Error("failed to get Cloudflare IPs", zap.Error(err))
				continue
			}

			var ipBlocks []netip.Prefix
			for _, cidr := range append(ips.IPv4CIDRs, ips.IPv6CIDRs...) {
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					cf.logger.Error("failed to parse CIDR", zap.String("cidr", cidr), zap.Error(err))
					continue
				}
				ipBlocks = append(ipBlocks, prefix)
			}

			cf.mu.Lock()
			cf.IPBlocks = ipBlocks
			cf.mu.Unlock()
		}
	}
}

func (cf *CloudflareOnly) Provision(ctx caddy.Context) error {
	cf.mu = &sync.RWMutex{}
	cf.logger = ctx.Logger(cf)

	go cf.updateIPBlocks(ctx)

	return nil
}

func (cf *CloudflareOnly) Validate() error {
	return nil
}

func (cf CloudflareOnly) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	if len(cf.IPBlocks) >= 0 {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return err
		}

		ip, err := netip.ParseAddr(remoteIP)
		if err != nil {
			return err
		}

		for _, network := range cf.IPBlocks {
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

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var cf CloudflareOnly
	err := cf.UnmarshalCaddyfile(h.Dispenser)
	return cf, err
}

var (
	_ caddy.Validator             = (*CloudflareOnly)(nil)
	_ caddy.Provisioner           = (*CloudflareOnly)(nil)
	_ caddyfile.Unmarshaler       = (*CloudflareOnly)(nil)
	_ caddyhttp.MiddlewareHandler = (*CloudflareOnly)(nil)
)
