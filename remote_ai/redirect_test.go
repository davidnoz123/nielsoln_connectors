package main

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

// scripted is a transport that answers with lines decided in advance.
//
// STRICT. It models exactly what hello() uses, sendRaw and readLine, and
// refuses to invent anything: a readLine past the end is io.EOF rather than
// a blank line, because a fake that quietly keeps answering lets a test pass
// while the code loops.
type scripted struct {
	lines []string
	at    int
	sent  [][]byte
}

func (s *scripted) sendRaw(raw []byte) error {
	s.sent = append(s.sent, append([]byte(nil), raw...))
	return nil
}

func (s *scripted) readLine() ([]byte, error) {
	if s.at >= len(s.lines) {
		return nil, io.EOF
	}
	line := s.lines[s.at]
	s.at++
	return []byte(line), nil
}

func (s *scripted) Close() error { return nil }

func notHere(seat, message string) string {
	body := map[string]any{"id": 0, "ok": false,
		"error": map[string]any{"code": "not_here", "message": message}}
	if seat != "" {
		body["error"].(map[string]any)["seat"] = seat
	}
	raw, _ := json.Marshal(body)
	return string(raw)
}

// THE CONTROL. Every refusal below would also be satisfied by a hello() that
// rejected everything, so the one case that must work comes first.
func TestATokenIsSentToTheBridgeThatHoldsIt(t *testing.T) {
	c := &connector{token: "AAAA1111", root: t.TempDir()}
	tr := &scripted{lines: []string{notHere("seat1", "lives elsewhere")}}

	err := c.hello(tr)

	var rd redirected
	if !errors.As(err, &rd) {
		t.Fatalf("wanted a redirect, got %v", err)
	}
	if rd.seat != "seat1" {
		t.Fatalf("sent to %q, wanted seat1", rd.seat)
	}
	if c.seat != "seat1" {
		t.Fatalf("the seat was not remembered: %q", c.seat)
	}
	if c.path() != "/bridge/seat1" {
		t.Fatalf("the next dial would be %q", c.path())
	}
	// The hello it sent must carry the token, since that is the only thing
	// the resolver has to work from.
	if len(tr.sent) != 1 || !strings.Contains(string(tr.sent[0]), "AAAA1111") {
		t.Fatalf("the hello did not carry the token: %q", tr.sent)
	}
}

func TestNoSeatMeansThisCopyIsFinished(t *testing.T) {
	c := &connector{token: "ZZZZ9999", root: t.TempDir()}
	tr := &scripted{lines: []string{
		notHere("", "This connector is out of date. Download a new one.")}}

	err := c.hello(tr)

	var rf refused
	if !errors.As(err, &rf) {
		t.Fatalf("wanted a refusal, got %v", err)
	}
	// The server's own words, unwrapped: they are the participant's
	// instructions and this connector has nothing to add to them.
	if !strings.Contains(rf.why, "Download a new one") {
		t.Fatalf("the server's message was not passed through: %q", rf.why)
	}
	if c.seat != "" {
		t.Fatalf("a refusal must not set a seat, got %q", c.seat)
	}
}

// A NAME, NEVER A URL. Anything that could carry a slash, a dot segment or a
// scheme would be asking this to dial somewhere else entirely.
func TestAnUnusableBridgeNameIsNotFollowed(t *testing.T) {
	for _, bad := range []string{
		"../bridge", "a/b", "a..b", "a%2fb", "wss://elsewhere/bridge",
		"Seat1", "1seat", "", "a_b", strings.Repeat("a", 40),
	} {
		c := &connector{token: "AAAA1111", root: t.TempDir()}
		tr := &scripted{lines: []string{notHere(bad, "lives elsewhere")}}

		err := c.hello(tr)

		var rf refused
		if !errors.As(err, &rf) {
			t.Fatalf("%q was not refused, got %v", bad, err)
		}
		if c.seat != "" {
			t.Fatalf("%q was remembered as a seat: %q", bad, c.seat)
		}
	}
}

// ONE REDIRECT, EVER. A second means the server is pointing in a circle, and
// following it would hide that behind a retry.
func TestASecondRedirectIsALoopAndIsNotFollowed(t *testing.T) {
	c := &connector{token: "AAAA1111", root: t.TempDir(), seat: "seat1"}
	tr := &scripted{lines: []string{notHere("seat2", "lives elsewhere")}}

	err := c.hello(tr)

	var rf refused
	if !errors.As(err, &rf) {
		t.Fatalf("wanted a refusal, got %v", err)
	}
	if !strings.Contains(rf.why, "loop") {
		t.Fatalf("the reason did not name the loop: %q", rf.why)
	}
	if c.seat != "seat1" {
		t.Fatalf("the remembered seat was overwritten: %q", c.seat)
	}
}

// internal is NOT a refusal. The resolver could not check the code, which is
// worth retrying, and telling the two apart is the whole reason it answers
// with two different codes.
func TestAServerThatCouldNotCheckIsRetried(t *testing.T) {
	c := &connector{token: "AAAA1111", root: t.TempDir()}
	raw, _ := json.Marshal(map[string]any{"id": 0, "ok": false,
		"error": map[string]any{"code": "internal",
			"message": "could not check your code just now"}})
	tr := &scripted{lines: []string{string(raw)}}

	err := c.hello(tr)

	if err == nil {
		t.Fatal("wanted an error")
	}
	var rf refused
	if errors.As(err, &rf) {
		t.Fatal("internal must not stop the connector: it is retryable")
	}
	var rd redirected
	if errors.As(err, &rd) {
		t.Fatal("internal is not a redirect")
	}
	if c.seat != "" {
		t.Fatalf("internal must not set a seat, got %q", c.seat)
	}
}

// And the ordinary case still works, so none of the above has broken the
// path every session takes.
func TestAnAcceptedHelloStillEstablishesTheSession(t *testing.T) {
	c := &connector{token: "AAAA1111", root: t.TempDir()}
	raw, _ := json.Marshal(map[string]any{"id": 0, "ok": true,
		"result": map[string]any{"session": "s1", "resumed": false}})
	tr := &scripted{lines: []string{string(raw)}}

	if err := c.hello(tr); err != nil {
		t.Fatalf("an accepted hello returned %v", err)
	}
	if c.session != "s1" {
		t.Fatalf("the session was not recorded: %q", c.session)
	}
	if c.seat != "" {
		t.Fatalf("an accepted hello must not invent a seat: %q", c.seat)
	}
}
