package xray

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	cfg "github.com/vasyza/vless-reality-tor-vpn/internal/config"
)

var (
	xrayConfigWrites = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vpnd",
		Name:      "xray_config_writes_total",
		Help:      "Number of times Xray config was generated/written",
	})
)

// MakeConfig builds an Xray JSON config from input parameters and users.
func MakeConfig(conf cfg.Config, users []uuid.UUID) (Root, error) {
	port := 443
	listenHost := conf.Xray.Listen
	if i := strings.LastIndex(listenHost, ":"); i >= 0 {
		p := listenHost[i+1:]
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	in := Inbound{
		Tag:      "vless-in",
		Port:     port,
		Listen:   "0.0.0.0",
		Protocol: "vless",
		Settings: InboundSettings{
			Clients:    []User{},
			Decryption: "none",
		},
		StreamSettings: StreamSettings{
			Network:  "tcp",
			Security: "reality",
			RealitySettings: &RealitySettings{
				Dest:        conf.Xray.Dest,
				ServerNames: conf.Xray.ServerNames,
				PrivateKey:  conf.Xray.PrivateKey,
				ShortIds:    conf.Xray.ShortIDs,
			},
		},
		Sniffing: &Sniffing{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}
	for _, u := range users {
		in.Settings.Clients = append(in.Settings.Clients, User{
			ID:         u.String(),
			Flow:       "xtls-rprx-vision",
		})
	}

	outbounds := []Outbound{
		{Tag: "direct", Protocol: "freedom"},
	}
	route := &Routing{DomainStrategy: "AsIs", Rules: []RoutingRule{}}

	if conf.Xray.Egress == "tor" {
		host, portS := splitHostPortDefault(conf.Tor.Socks, "9050")
		portI, _ := strconv.Atoi(portS)
		outbounds = append(outbounds, Outbound{
			Tag:      "tor",
			Protocol: "socks",
			Settings: SocksSettings{Servers: []SocksServer{{Address: host, Port: portI}}},
		})
		route.Rules = append(route.Rules, RoutingRule{Type: "field", OutboundTag: "tor", Port: "0-65535"})
	}

	root := Root{
		Log:       &Log{LogLevel: conf.Xray.LogLevel},
		Inbounds:  []Inbound{in},
		Outbounds: outbounds,
		Routing:   route,
	}
	return root, nil
}

func splitHostPortDefault(addr, defaultPort string) (string, string) {
	if strings.Contains(addr, ":") {
		parts := strings.Split(addr, ":")
		if len(parts) > 1 && parts[len(parts)-1] != "" {
			return strings.Join(parts[:len(parts)-1], ":"), parts[len(parts)-1]
		}
	}
	return addr, defaultPort
}

// Write writes cfg as JSON to path.
func Write(path string, cfg Root) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}
	xrayConfigWrites.Inc()
	return nil
}

func dir(p string) string {
	i := strings.LastIndex(p, "/")
	if i <= 0 {
		return "."
	}
	return p[:i]
}

func (c Root) Pretty() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func ValidateRealityFields(conf cfg.Config) error {
	if conf.Xray.PrivateKey == "" {
		return fmt.Errorf("xray.private_key is empty: generate one via `xray x25519` and set it in config")
	}
	if len(conf.Xray.ShortIDs) == 0 {
		return fmt.Errorf("xray.short_ids must contain at least one value (8-16 hex chars)")
	}
	return nil
}
