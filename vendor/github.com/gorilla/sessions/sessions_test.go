package sessions

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"testing"
)

type ResponseRecorder struct {
	Code		int
	HeaderMap	http.Header
	Body		*bytes.Buffer
	Flushed		bool
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap:	make(http.Header),
		Body:		new(bytes.Buffer),
	}
}

const DefaultRemoteAddr = "1.2.3.4"

func (rw *ResponseRecorder) Header() http.Header {
	return rw.HeaderMap
}

func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	if rw.Code == 0 {
		rw.Code = http.StatusOK
	}
	return len(buf), nil
}

func (rw *ResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

func (rw *ResponseRecorder) Flush() {
	rw.Flushed = true
}

type FlashMessage struct {
	Type	int
	Message	string
}

func TestFlashes(t *testing.T) {
	var req *http.Request
	var rsp *ResponseRecorder
	var hdr http.Header
	var err error
	var ok bool
	var cookies []string
	var session *Session
	var flashes []interface{}

	store := NewCookieStore([]byte("secret-key"))

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()

	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected empty flashes; Got %v", flashes)
	}

	session.AddFlash("foo")
	session.AddFlash("bar")

	session.AddFlash("baz", "custom_key")

	if err = Save(req, rsp); err != nil {
		t.Fatalf("Error saving session: %v", err)
	}
	hdr = rsp.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()

	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	flashes = session.Flashes()
	if len(flashes) != 2 {
		t.Fatalf("Expected flashes; Got %v", flashes)
	}
	if flashes[0] != "foo" || flashes[1] != "bar" {
		t.Errorf("Expected foo,bar; Got %v", flashes)
	}
	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}

	flashes = session.Flashes("custom_key")
	if len(flashes) != 1 {
		t.Errorf("Expected flashes; Got %v", flashes)
	} else if flashes[0] != "baz" {
		t.Errorf("Expected baz; Got %v", flashes)
	}
	flashes = session.Flashes("custom_key")
	if len(flashes) != 0 {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()

	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected empty flashes; Got %v", flashes)
	}

	session.AddFlash(&FlashMessage{42, "foo"})

	if err = Save(req, rsp); err != nil {
		t.Fatalf("Error saving session: %v", err)
	}
	hdr = rsp.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()

	if session, err = store.Get(req, "session-key"); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	flashes = session.Flashes()
	if len(flashes) != 1 {
		t.Fatalf("Expected flashes; Got %v", flashes)
	}
	custom := flashes[0].(FlashMessage)
	if custom.Type != 42 || custom.Message != "foo" {
		t.Errorf("Expected %#v, got %#v", FlashMessage{42, "foo"}, custom)
	}
}

func init() {
	gob.Register(FlashMessage{})
}
