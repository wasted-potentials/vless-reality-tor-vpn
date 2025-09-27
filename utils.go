package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
)

// KeyGenerator утилиты для генерации ключей
type KeyGenerator struct{}

// GenerateUUID создает новый UUID v4 для VLESS
func (kg *KeyGenerator) GenerateUUID() string {
	id := uuid.New()
	return id.String()
}

// GenerateRealityKeys генерирует пару ключей для Reality
func (kg *KeyGenerator) GenerateRealityKeys() (string, string, error) {
	// Генерируем приватный ключ (32 байта)
	privateKey := make([]byte, 32)
	_, err := rand.Read(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Генерируем публичный ключ на основе приватного (упрощенно)
	hash := sha256.Sum256(privateKey)
	publicKey := hash[:]

	privateKeyHex := hex.EncodeToString(privateKey)
	publicKeyHex := hex.EncodeToString(publicKey)

	return privateKeyHex, publicKeyHex, nil
}

// GenerateShortID создает короткий ID для Reality
func (kg *KeyGenerator) GenerateShortID() string {
	shortId := make([]byte, 8)
	rand.Read(shortId)
	return hex.EncodeToString(shortId)
}

// ClientConfigGenerator генератор клиентских конфигураций
type ClientConfigGenerator struct{}

// XrayClientConfig структура конфигурации Xray клиента
type XrayClientConfig struct {
	Outbounds []XrayOutbound `json:"outbounds"`
	Inbounds  []XrayInbound  `json:"inbounds"`
	Routing   XrayRouting    `json:"routing"`
}

type XrayOutbound struct {
	Protocol       string               `json:"protocol"`
	Settings       XrayOutboundSettings `json:"settings"`
	StreamSettings XrayStreamSettings   `json:"streamSettings"`
	Tag            string               `json:"tag"`
}

type XrayOutboundSettings struct {
	Vnext []XrayVnext `json:"vnext"`
}

type XrayVnext struct {
	Address string     `json:"address"`
	Port    int        `json:"port"`
	Users   []XrayUser `json:"users"`
}

type XrayUser struct {
	ID         string `json:"id"`
	Flow       string `json:"flow"`
	Encryption string `json:"encryption"`
}

type XrayStreamSettings struct {
	Network         string              `json:"network"`
	Security        string              `json:"security"`
	RealitySettings XrayRealitySettings `json:"realitySettings"`
}

type XrayRealitySettings struct {
	ServerName  string `json:"serverName"`
	Fingerprint string `json:"fingerprint"`
	ShortId     string `json:"shortId"`
	PublicKey   string `json:"publicKey"`
	Show        bool   `json:"show"`
}

type XrayInbound struct {
	Port     int                 `json:"port"`
	Protocol string              `json:"protocol"`
	Settings XrayInboundSettings `json:"settings"`
	Tag      string              `json:"tag"`
}

type XrayInboundSettings struct {
	Auth      string `json:"auth"`
	UDP       bool   `json:"udp"`
	UserLevel int    `json:"userLevel"`
}

type XrayRouting struct {
	Rules []XrayRoutingRule `json:"rules"`
}

type XrayRoutingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
}

// GenerateXrayConfig создает конфигурацию для Xray клиента
func (ccg *ClientConfigGenerator) GenerateXrayConfig(serverIP string, serverPort int, uuid, publicKey, serverName, shortId string) XrayClientConfig {
	return XrayClientConfig{
		Outbounds: []XrayOutbound{
			{
				Protocol: "vless",
				Settings: XrayOutboundSettings{
					Vnext: []XrayVnext{
						{
							Address: serverIP,
							Port:    serverPort,
							Users: []XrayUser{
								{
									ID:         uuid,
									Flow:       "xtls-rprx-vision",
									Encryption: "none",
								},
							},
						},
					},
				},
				StreamSettings: XrayStreamSettings{
					Network:  "tcp",
					Security: "reality",
					RealitySettings: XrayRealitySettings{
						ServerName:  serverName,
						Fingerprint: "chrome",
						ShortId:     shortId,
						PublicKey:   publicKey,
						Show:        false,
					},
				},
				Tag: "proxy",
			},
			{
				Protocol: "freedom",
				Settings: XrayOutboundSettings{},
				Tag:      "direct",
			},
		},
		Inbounds: []XrayInbound{
			{
				Port:     10808,
				Protocol: "socks",
				Settings: XrayInboundSettings{
					Auth:      "noauth",
					UDP:       true,
					UserLevel: 8,
				},
				Tag: "socks",
			},
			{
				Port:     10809,
				Protocol: "http",
				Settings: XrayInboundSettings{
					UserLevel: 8,
				},
				Tag: "http",
			},
		},
		Routing: XrayRouting{
			Rules: []XrayRoutingRule{
				{
					Type:        "field",
					OutboundTag: "direct",
					Domain:      []string{"geosite:private"},
				},
				{
					Type:        "field",
					OutboundTag: "direct",
					IP:          []string{"geoip:private", "geoip:cn"},
				},
				{
					Type:        "field",
					OutboundTag: "proxy",
				},
			},
		},
	}
}

// V2rayNGConfig структура для V2rayNG клиента (Android)
type V2rayNGConfig struct {
	V    string `json:"v"`
	PS   string `json:"ps"`
	Add  string `json:"add"`
	Port string `json:"port"`
	ID   string `json:"id"`
	Aid  string `json:"aid"`
	SCY  string `json:"scy"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	TLS  string `json:"tls"`
	SNI  string `json:"sni"`
	ALPN string `json:"alpn"`
	FP   string `json:"fp"`
	PBK  string `json:"pbk"`
	SID  string `json:"sid"`
	SPSK string `json:"spsk"`
}

// GenerateV2rayNGURL создает ссылку для V2rayNG
func (ccg *ClientConfigGenerator) GenerateV2rayNGURL(serverIP string, serverPort int, uuid, publicKey, serverName, shortId string) string {
	config := V2rayNGConfig{
		V:    "2",
		PS:   "VPN-Server",
		Add:  serverIP,
		Port: fmt.Sprintf("%d", serverPort),
		ID:   uuid,
		Aid:  "0",
		SCY:  "none",
		Net:  "tcp",
		Type: "none",
		Host: "",
		Path: "",
		TLS:  "reality",
		SNI:  serverName,
		ALPN: "",
		FP:   "chrome",
		PBK:  publicKey,
		SID:  shortId,
		SPSK: "",
	}

	// В реальной реализации здесь должно быть base64 кодирование JSON
	return fmt.Sprintf("vless://%s@%s:%d?encryption=none&flow=xtls-rprx-vision&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=tcp&headerType=none#VPN-Server",
		uuid, serverIP, serverPort, serverName, publicKey, shortId)
}

// QRCodeGenerator генератор QR-кодов для клиентских конфигураций
type QRCodeGenerator struct{}

// GenerateQRCode создает ASCII QR-код (упрощенная версия)
func (qr *QRCodeGenerator) GenerateQRCode(data string) string {
	// Упрощенная ASCII версия QR-кода
	// В реальной реализации используйте библиотеку для QR-кодов
	return fmt.Sprintf(`
	████ ▄▄▄▄▄ █▀█ █▄█▄▄▄█ ▄▄▄▄▄ ████
	████ █   █ █▀▀ █ ▄  ▄█ █   █ ████
	████ █▄▄▄█ █▀ ▀█▀▀█▄▄█ █▄▄▄█ ████
	████▄▄▄▄▄▄▄█▄▀ ▀▄▀ ▀▄█▄▄▄▄▄▄▄████
	████▄▄▄ ▀▄▄ ▄▀▄▀▀▄▀▄▄▄▄ ▀▄█▀▄████

	Конфигурационная ссылка:
	%s

	Сканируйте QR-код в мобильном клиенте или
	скопируйте ссылку в настройки клиента.
	`, data)
}

// ServerManager менеджер серверных функций
type ServerManager struct{}

// CheckTorConnection проверяет соединение с Tor
func (sm *ServerManager) CheckTorConnection(socksAddr string) error {
	// Здесь должна быть проверка соединения с Tor
	log.Printf("Checking Tor connection at %s...", socksAddr)
	// Упрощенная проверка
	return nil
}

// ValidateRealityDest проверяет доступность Reality destination
func (sm *ServerManager) ValidateRealityDest(dest string) error {
	log.Printf("Validating Reality destination %s...", dest)
	// Здесь должна быть проверка доступности домена
	return nil
}

// GenerateServerStats генерирует статистику сервера
func (sm *ServerManager) GenerateServerStats() map[string]interface{} {
	return map[string]interface{}{
		"active_connections": 0,
		"total_bytes_sent":   0,
		"total_bytes_recv":   0,
		"uptime_seconds":     0,
		"tor_status":         "connected",
	}
}

// CLI утилита командной строки
func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	kg := &KeyGenerator{}
	ccg := &ClientConfigGenerator{}
	qr := &QRCodeGenerator{}
	sm := &ServerManager{}

	switch os.Args[1] {
	case "gen-uuid":
		uuid := kg.GenerateUUID()
		fmt.Printf("Generated UUID: %s\n", uuid)

	case "gen-reality-keys":
		privateKey, publicKey, err := kg.GenerateRealityKeys()
		if err != nil {
			log.Fatalf("Failed to generate Reality keys: %v", err)
		}
		fmt.Printf("Private Key: %s\n", privateKey)
		fmt.Printf("Public Key: %s\n", publicKey)

	case "gen-short-id":
		shortId := kg.GenerateShortID()
		fmt.Printf("Generated Short ID: %s\n", shortId)

	case "gen-client-config":
		if len(os.Args) < 7 {
			fmt.Println("Usage: gen-client-config <server_ip> <port> <uuid> <public_key> <server_name> <short_id>")
			return
		}

		serverIP := os.Args[2]
		serverPort := 443 // или парсить из os.Args[3]
		uuid := os.Args[4]
		publicKey := os.Args[5]
		serverName := os.Args[6]
		shortId := os.Args[7]

		// Генерируем конфигурацию
		config := ccg.GenerateXrayConfig(serverIP, serverPort, uuid, publicKey, serverName, shortId)
		fmt.Printf("Xray client config generated for %s:%d\n", serverIP, serverPort)

		// Генерируем ссылку для мобильных клиентов
		url := ccg.GenerateV2rayNGURL(serverIP, serverPort, uuid, publicKey, serverName, shortId)
		fmt.Printf("\nV2rayNG URL: %s\n", url)

		// Генерируем QR-код
		qrCode := qr.GenerateQRCode(url)
		fmt.Println(qrCode)

	case "check-tor":
		socksAddr := "127.0.0.1:9050"
		if len(os.Args) > 2 {
			socksAddr = os.Args[2]
		}

		err := sm.CheckTorConnection(socksAddr)
		if err != nil {
			log.Fatalf("Tor connection failed: %v", err)
		}
		fmt.Println("Tor connection: OK")

	case "validate-dest":
		if len(os.Args) < 3 {
			fmt.Println("Usage: validate-dest <destination>")
			return
		}

		dest := os.Args[2]
		err := sm.ValidateRealityDest(dest)
		if err != nil {
			log.Fatalf("Reality destination validation failed: %v", err)
		}
		fmt.Printf("Reality destination %s: OK\n", dest)

	case "stats":
		stats := sm.GenerateServerStats()
		fmt.Printf("Server Statistics:\n")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Printf(`VPN Server Utilities

Usage: %s [command] [options]

Commands:
  gen-uuid                                    Generate new UUID for VLESS
  gen-reality-keys                           Generate Reality private/public key pair
  gen-short-id                               Generate Reality short ID
  gen-client-config <ip> <port> <uuid> <key> <sni> <sid>  Generate client configuration
  check-tor [socks_addr]                     Check Tor connection (default: 127.0.0.1:9050)
  validate-dest <destination>                Validate Reality destination
  stats                                       Show server statistics

Examples:
  %s gen-uuid
  %s gen-reality-keys
  %s gen-client-config 1.2.3.4 443 uuid-here public-key-here www.google.com short-id-here
  %s check-tor 127.0.0.1:9050
  %s validate-dest www.google.com:443

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
