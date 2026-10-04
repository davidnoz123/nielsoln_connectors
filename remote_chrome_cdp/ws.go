// ws.go -- the smallest websocket client that can carry CDP.
//
// WHY NOT THE ONE IN remote_ai. That client carries TLS, permessage-deflate
// and a handshake hardened against captive portals, because it crosses the
// internet through Caddy. This one talks to 127.0.0.1 and nothing else, so
// every one of those costs buys nothing here. When the bridge transport
// lands, remote_ai's dial() is the thing to lift, not this.
//
// WHY NOT A LIBRARY. remote_ai's go.mod has no require block at all, and
// that is the convention worth keeping: `go run <module>@<sha>` on somebody
// else's machine should fetch one thing, not a dependency tree they did not
// choose and cannot audit.
package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type wsConn struct {
	conn net.Conn
	r    *bufio.Reader
}

// wsDial opens a websocket to rawURL, which must be ws:// on loopback.
//
// The scheme is checked rather than assumed. A wss:// here would silently
// skip TLS and send the frames in the clear, and the only symptom would be
// a connection that works perfectly.
func wsDial(rawURL string, timeout time.Duration) (*wsConn, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "ws" {
		return nil, fmt.Errorf("ws.go speaks ws:// only, not %q", u.Scheme)
	}
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		conn.Close()
		return nil, fmt.Errorf("no randomness for the websocket key: %w", err)
	}
	key := base64.StdEncoding.EncodeToString(raw)
	path := u.RequestURI()
	req := strings.Join([]string{
		"GET " + path + " HTTP/1.1",
		"Host: " + u.Host,
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Key: " + key,
		"Sec-WebSocket-Version: 13",
	}, "\r\n") + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}
	r := bufio.NewReaderSize(conn, 1<<16)
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("no HTTP response to the upgrade: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("chrome refused the upgrade: %s", resp.Status)
	}
	// Checked, not trusted. An intermediary answering 101 without the right
	// accept key gives a socket that looks open and never speaks CDP, which
	// is indistinguishable from a hang.
	sum := sha1.Sum([]byte(key + wsGUID))
	if resp.Header.Get("Sec-WebSocket-Accept") !=
		base64.StdEncoding.EncodeToString(sum[:]) {
		conn.Close()
		return nil, errors.New("Sec-WebSocket-Accept did not match the key we sent")
	}
	return &wsConn{conn: conn, r: r}, nil
}

func (w *wsConn) Close() error { return w.conn.Close() }

// writeText sends one unfragmented masked text frame.
//
// MASKED, always. A client that sends unmasked frames is a protocol error
// and Chrome closes the connection without explaining why.
func (w *wsConn) writeText(payload []byte) error {
	var head []byte
	n := len(payload)
	switch {
	case n < 126:
		head = []byte{0x81, byte(0x80 | n)}
	case n < 1<<16:
		head = []byte{0x81, 0x80 | 126, 0, 0}
		binary.BigEndian.PutUint16(head[2:], uint16(n))
	default:
		head = make([]byte, 10)
		head[0], head[1] = 0x81, 0x80|127
		binary.BigEndian.PutUint64(head[2:], uint64(n))
	}
	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return err
	}
	out := make([]byte, 0, len(head)+4+n)
	out = append(out, head...)
	out = append(out, mask[:]...)
	for i := 0; i < n; i++ {
		out = append(out, payload[i]^mask[i%4])
	}
	_, err := w.conn.Write(out)
	return err
}

// readText returns the next text message, answering pings and skipping
// anything else. CDP never fragments its replies in practice, but a reply
// bigger than 64KB arrives in continuation frames, so they are joined
// rather than assumed away.
func (w *wsConn) readText(deadline time.Time) ([]byte, error) {
	var msg []byte
	for {
		if err := w.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
		var h [2]byte
		if _, err := io.ReadFull(w.r, h[:]); err != nil {
			return nil, err
		}
		fin := h[0]&0x80 != 0
		opcode := h[0] & 0x0f
		n := int(h[1] & 0x7f)
		switch n {
		case 126:
			var e [2]byte
			if _, err := io.ReadFull(w.r, e[:]); err != nil {
				return nil, err
			}
			n = int(binary.BigEndian.Uint16(e[:]))
		case 127:
			var e [8]byte
			if _, err := io.ReadFull(w.r, e[:]); err != nil {
				return nil, err
			}
			n = int(binary.BigEndian.Uint64(e[:]))
		}
		// A server frame is never masked; if one is, the stream is not what
		// we think it is and reading on would return rubbish.
		if h[1]&0x80 != 0 {
			return nil, errors.New("chrome sent a masked frame, which is a protocol error")
		}
		body := make([]byte, n)
		if _, err := io.ReadFull(w.r, body); err != nil {
			return nil, err
		}
		switch opcode {
		case 0x8:
			return nil, errors.New("chrome closed the connection")
		case 0x9: // ping
			pong := append([]byte{}, body...)
			if err := w.writeFrame(0xA, pong); err != nil {
				return nil, err
			}
			continue
		case 0xA: // pong
			continue
		case 0x1, 0x0: // text, continuation
			msg = append(msg, body...)
			if fin {
				return msg, nil
			}
		default:
			continue
		}
	}
}

func (w *wsConn) writeFrame(opcode byte, payload []byte) error {
	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return err
	}
	out := []byte{0x80 | opcode, byte(0x80 | len(payload))}
	out = append(out, mask[:]...)
	for i := range payload {
		out = append(out, payload[i]^mask[i%4])
	}
	_, err := w.conn.Write(out)
	return err
}
