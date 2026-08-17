package transport

import (
	"errors"
	"fmt"
	"net"
	"os"
)

const maxTokenLength = 128

// OpenFromEnv connects to the loopback UDP socket owned by the external
// libopenconnect helper. UDP preserves one IP packet per datagram.
func OpenFromEnv() (net.Conn, error) {
	peer := os.Getenv("VPN_UDP_PEER")
	if peer == "" {
		return nil, errors.New("VPN_UDP_PEER is required on Windows")
	}
	token := os.Getenv("VPN_UDP_TOKEN")
	if token == "" {
		return nil, errors.New("VPN_UDP_TOKEN is required when VPN_UDP_PEER is set")
	}
	if len(token) > maxTokenLength {
		return nil, fmt.Errorf("VPN_UDP_TOKEN exceeds %d bytes", maxTokenLength)
	}

	peerAddr, err := net.ResolveUDPAddr("udp4", peer)
	if err != nil {
		return nil, fmt.Errorf("resolve VPN_UDP_PEER %q: %w", peer, err)
	}
	if !peerAddr.IP.IsLoopback() || peerAddr.Port == 0 {
		return nil, fmt.Errorf("VPN_UDP_PEER must be a loopback address with a non-zero port: %q", peer)
	}

	conn, err := net.DialUDP("udp4", nil, peerAddr)
	if err != nil {
		return nil, fmt.Errorf("dial VPN_UDP_PEER %q: %w", peer, err)
	}
	if _, err := conn.Write([]byte(token)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake VPN_UDP_PEER %q: %w", peer, err)
	}
	return conn, nil
}
