
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/curve25519"
)

func main() {
	// If xray binary available, prefer its generator to avoid format mismatches.
	if bin := findXray(); bin != "" {
		fmt.Fprintf(os.Stderr, "Using %s x25519 key generator...\n", bin)
		out, err := exec.Command(bin, "x25519").CombinedOutput()
		if err == nil && len(out) > 0 {
			fmt.Print(string(out))
			return
		}
	}

	// Fallback: local generator (hex-encoded).
	priv, pub := genX25519Hex()
	fmt.Println("PrivateKey (hex):", priv)
	fmt.Println("PublicKey  (hex):", pub)
	fmt.Println("\nNOTE: If Xray rejects the format, run `xray x25519` on the server and copy the PrivateKey into config.xray.private_key")
}

func findXray() string {
	path := os.Getenv("XRAY")
	if path != "" { return path }
	path, _ = exec.LookPath("xray")
	return path
}

func genX25519Hex() (string, string) {
	var sk [32]byte
	if _, err := rand.Read(sk[:]); err != nil {
		panic(err)
	}
	clamp(&sk)
	var pk [32]byte
	curve25519.ScalarBaseMult(&pk, &sk)
	return hex.EncodeToString(sk[:]), hex.EncodeToString(pk[:])
}

func clamp(s *[32]byte) {
	s[0] &= 248
	s[31] &= 127
	s[31] |= 64
}

// helper to parse `xray x25519` output if needed
func parseXray(out []byte) (string, string) {
	lines := bytes.Split(out, []byte("\n"))
	var priv, pub string
	for _, l := range lines {
		s := string(l)
		if strings.Contains(s, "Private key:") {
			priv = strings.TrimSpace(strings.TrimPrefix(s, "Private key:"))
		}
		if strings.Contains(s, "Public key:") {
			pub = strings.TrimSpace(strings.TrimPrefix(s, "Public key:"))
		}
	}
	return priv, pub
}
