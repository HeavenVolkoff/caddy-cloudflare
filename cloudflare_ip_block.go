package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

type CloudflareIPs struct {
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
}

type CloudflareIPBlock struct {
	ips    []netip.Prefix
	lock   sync.RWMutex
	logger *zap.Logger
}

func (cf *CloudflareIPBlock) updateIPBlocks(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		ips, err := FetchCloudflareIPs(ctx)
		if err != nil {
			cf.logger.Error("failed to get Cloudflare IPs", zap.Error(err))
		} else {
			var ipBlocks []netip.Prefix
			for _, cidr := range append(ips.IPv4CIDRs, ips.IPv6CIDRs...) {
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					cf.logger.Error("failed to parse Cloudflare CIDR",
						zap.String("cidr", cidr), zap.Error(err))
					continue
				}
				ipBlocks = append(ipBlocks, prefix)
			}

			cf.lock.Lock()
			cf.ips = ipBlocks
			cf.logger.Info("updated Cloudflare IP blocks", zap.Any("ip_blocks", ipBlocks))
			cf.lock.Unlock()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
}

var (
	cf   *CloudflareIPBlock
	once sync.Once
)

func GetCloudflareIpBlock(ctx caddy.Context) *CloudflareIPBlock {
	once.Do(func() {
		cf = &CloudflareIPBlock{logger: ctx.Logger()}
		go cf.updateIPBlocks(ctx)
	})
	return cf
}

func FetchCloudflareIPs(ctx context.Context) (*CloudflareIPs, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.cloudflare.com/client/v4/ips", nil)
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
