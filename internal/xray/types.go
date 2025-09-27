
package xray

// Partial config models sufficient for generation.
// These mirror Xray JSON schema in a minimal way.

type User struct {
	ID         string `json:"id"`
	Email      string `json:"email,omitempty"`
	Flow       string `json:"flow,omitempty"`       // e.g., "xtls-rprx-vision"
	Encryption string `json:"encryption,omitempty"` // for VLESS clients only "none"
	Level      int    `json:"level,omitempty"`
}

type InboundSettings struct {
	Clients    []User `json:"clients"`
	Decryption string `json:"decryption"` // "none"
}

type RealitySettings struct {
	Show        bool     `json:"show,omitempty"`
	Dest        string   `json:"dest"`
	Xver        int      `json:"xver,omitempty"`
	ServerNames []string `json:"serverNames"`
	PrivateKey  string   `json:"privateKey"`
	ShortIds    []string `json:"shortIds"`
	SpiderX     string   `json:"spiderX,omitempty"`
}

type StreamSettings struct {
	Network         string          `json:"network"`           // "tcp"
	Security        string          `json:"security"`          // "reality"
	RealitySettings *RealitySettings `json:"realitySettings"`
}

type Inbound struct {
	Tag           string           `json:"tag,omitempty"`
	Port          int              `json:"port"`
	Listen        string           `json:"listen,omitempty"`
	Protocol      string           `json:"protocol"` // "vless"
	Settings      InboundSettings  `json:"settings"`
	StreamSettings StreamSettings  `json:"streamSettings"`
	Sniffing      *Sniffing        `json:"sniffing,omitempty"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

type SocksServer struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type SocksSettings struct {
	Servers []SocksServer `json:"servers"`
}

type Outbound struct {
	Tag      string       `json:"tag,omitempty"`
	Protocol string       `json:"protocol"`
	Settings interface{}  `json:"settings,omitempty"`
	// For VLESS/Trojan/SOCKS etc, per-protocol settings object
}

type RoutingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag"`
	IP          []string `json:"ip,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
}

type Routing struct {
	DomainStrategy string        `json:"domainStrategy,omitempty"`
	Rules          []RoutingRule `json:"rules"`
}

type Log struct {
	Access   string `json:"access,omitempty"`
	Error    string `json:"error,omitempty"`
	LogLevel string `json:"loglevel,omitempty"`
}

type Root struct {
	Log       *Log      `json:"log,omitempty"`
	Inbounds  []Inbound `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   *Routing  `json:"routing,omitempty"`
	Policy    interface{} `json:"policy,omitempty"`
	Stats     interface{} `json:"stats,omitempty"`
	API       interface{} `json:"api,omitempty"`
}
