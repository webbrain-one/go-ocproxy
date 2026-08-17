package transport

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestOpenFromEnvUsesConnectedUDP(t *testing.T) {
	t.Setenv("VPN_UDP_TOKEN", "test-token")
	listener, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	t.Setenv("VPN_UDP_PEER", listener.LocalAddr().String())

	conn, err := OpenFromEnv()
	if err != nil {
		t.Fatalf("OpenFromEnv: %v", err)
	}
	defer conn.Close()

	if err := listener.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	n, proxyAddr, err := listener.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read handshake: %v", err)
	}
	if got := string(buf[:n]); got != "test-token" {
		t.Fatalf("handshake = %q", got)
	}

	fromHelper := []byte{0x45, 0x00, 0x00, 0x14}
	if _, err := listener.WriteToUDP(fromHelper, proxyAddr); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	n, err = conn.Read(buf)
	if err != nil || string(buf[:n]) != string(fromHelper) {
		t.Fatalf("read = %v, %v", buf[:n], err)
	}

	toHelper := []byte{0x60, 0x00, 0x00, 0x00}
	if _, err := conn.Write(toHelper); err != nil {
		t.Fatal(err)
	}
	n, _, err = listener.ReadFromUDP(buf)
	if err != nil || string(buf[:n]) != string(toHelper) {
		t.Fatalf("write = %v, %v", buf[:n], err)
	}
}

func TestOpenFromEnvValidatesConfiguration(t *testing.T) {
	t.Setenv("VPN_UDP_PEER", "")
	if _, err := OpenFromEnv(); err == nil {
		t.Fatal("expected missing peer error")
	}

	t.Setenv("VPN_UDP_PEER", "192.0.2.1:21082")
	t.Setenv("VPN_UDP_TOKEN", "token")
	if _, err := OpenFromEnv(); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("expected loopback error, got %v", err)
	}

	t.Setenv("VPN_UDP_PEER", "127.0.0.1:21082")
	t.Setenv("VPN_UDP_TOKEN", "")
	if _, err := OpenFromEnv(); err == nil {
		t.Fatal("expected missing token error")
	}

	t.Setenv("VPN_UDP_TOKEN", strings.Repeat("x", maxTokenLength+1))
	if _, err := OpenFromEnv(); err == nil {
		t.Fatal("expected oversized token error")
	}
}
