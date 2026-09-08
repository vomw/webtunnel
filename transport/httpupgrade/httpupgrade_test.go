package httpupgrade

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
)

const payload = "HELLO"

type fakeConn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (c fakeConn) Read(p []byte) (int, error)  { return c.r.Read(p) }
func (c fakeConn) Write(p []byte) (int, error) { return c.w.Write(p) }

func TestClientKeepsBufferedPayload(t *testing.T) {
	var written bytes.Buffer
	conn := fakeConn{
		r: strings.NewReader("HTTP/1.1 101 Switching Protocols\r\nConnection: upgrade\r\nUpgrade: websocket\r\n\r\n" + payload),
		w: &written,
	}

	transport, err := NewHTTPUpgradeTransport(&Config{Path: "test", Host: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	upgraded, err := transport.Client(conn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(written.String(), "GET /test HTTP/1.1\r\n") {
		t.Errorf("unexpected request: %q", written.String())
	}
	assertPayload(t, upgraded)
}

func TestServerKeepsBufferedPayload(t *testing.T) {
	var written bytes.Buffer
	conn := fakeConn{
		r: strings.NewReader("GET /test HTTP/1.1\r\nHost: example.com\r\nConnection: upgrade\r\nUpgrade: websocket\r\n\r\n" + payload),
		w: &written,
	}

	transport, err := NewHTTPUpgradeTransport(&Config{Path: "test", Host: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	upgraded, err := transport.Server(conn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(written.String(), "HTTP/1.1 101 Switching Protocols\r\n") {
		t.Errorf("unexpected response: %q", written.String())
	}
	assertPayload(t, upgraded)
}

func assertPayload(t *testing.T, conn net.Conn) {
	t.Helper()
	buf := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("reading payload after handshake: %v", err)
	}
	if string(buf) != payload {
		t.Fatalf("got %q, want %q", buf, payload)
	}
}

func TestClientPayloadArrivingAfterHandshake(t *testing.T) {
	var written bytes.Buffer
	conn := fakeConn{
		r: io.MultiReader(
			strings.NewReader("HTTP/1.1 101 Switching Protocols\r\nConnection: upgrade\r\nUpgrade: websocket\r\n\r\n"),
			strings.NewReader(payload),
		),
		w: &written,
	}

	transport, err := NewHTTPUpgradeTransport(&Config{Path: "test", Host: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	upgraded, err := transport.Client(conn)
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.(wrappedConn).reader != conn {
		t.Error("handshake reader retained with an empty buffer")
	}
	assertPayload(t, upgraded)
}
