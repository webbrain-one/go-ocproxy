package netstack

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestRunWithWindowsUDPTransport(t *testing.T) {
	helper, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer helper.Close()

	tunnel, err := net.DialUDP("udp4", nil, helper.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer tunnel.Close()

	if _, err := tunnel.Write([]byte("test-token")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	if err := helper.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	_, goAddr, err := helper.ReadFromUDP(buf)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := New("10.0.0.1", 1500, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ns.Close()
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- ns.Run(ctx, tunnel) }()

	pkt := make([]byte, 20)
	pkt[0] = 0x45
	pkt[2], pkt[3] = 0, 20
	pkt[8], pkt[9] = 64, 6
	copy(pkt[12:16], []byte{10, 0, 0, 2})
	copy(pkt[16:20], []byte{10, 0, 0, 1})
	if _, err := helper.WriteToUDP(pkt, goAddr); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("Run exited before cancellation: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after cancellation")
	}
}
