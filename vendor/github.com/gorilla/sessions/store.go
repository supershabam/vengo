package sessions

import (
	"encoding/base32"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/supershabam/vengo/github.com/gorilla/securecookie"
)

type Store interface {
	Get(r *http.Request, name string) (*Session, error)

	New(r *http.Request, name string) (*Session, error)

	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}

func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	return &CookieStore{
		Codecs:	securecookie.CodecsFromPairs(keyPairs...),
		Options: &Options{
			Path:	"/",
			MaxAge:	86400 * 30,
		},
	}
}

type CookieStore struct {
	Codecs	[]securecookie.Codec
	Options	*Options
}

func (s *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	return GetRegistry(r).Get(s, name)
}

func (s *CookieStore) New(r *http.Request, name string) (*Session, error) {
	session := NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.Values,
			s.Codecs...)
		if err == nil {
			session.IsNew = false
		}
	}
	return session, err
}

func (s *CookieStore) Save(r *http.Request, w http.ResponseWriter,
	session *Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, NewCookie(session.Name(), encoded, session.Options))
	return nil
}

var fileMutex sync.RWMutex

func NewFilesystemStore(path string, keyPairs ...[]byte) *FilesystemStore {
	if path == "" {
		path = os.TempDir()
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	return &FilesystemStore{
		Codecs:	securecookie.CodecsFromPairs(keyPairs...),
		Options: &Options{
			Path:	"/",
			MaxAge:	86400 * 30,
		},
		path:	path,
	}
}

type FilesystemStore struct {
	Codecs	[]securecookie.Codec
	Options	*Options
	path	string
}

func (s *FilesystemStore) MaxLength(l int) {
	for _, c := range s.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(l)
		}
	}
}

func (s *FilesystemStore) Get(r *http.Request, name string) (*Session, error) {
	return GetRegistry(r).Get(s, name)
}

func (s *FilesystemStore) New(r *http.Request, name string) (*Session, error) {
	session := NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

func (s *FilesystemStore) Save(r *http.Request, w http.ResponseWriter,
	session *Session) error {
	if session.ID == "" {

		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *FilesystemStore) save(session *Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}
	filename := s.path + "session_" + session.ID
	fileMutex.Lock()
	defer fileMutex.Unlock()
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err = fp.Write([]byte(encoded)); err != nil {
		return err
	}
	fp.Close()
	return nil
}

func (s *FilesystemStore) load(session *Session) error {
	filename := s.path + "session_" + session.ID
	fp, err := os.OpenFile(filename, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	defer fp.Close()
	var fdata []byte
	buf := make([]byte, 128)
	for {
		var n int
		n, err = fp.Read(buf[0:])
		fdata = append(fdata, buf[0:n]...)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	if err = securecookie.DecodeMulti(session.Name(), string(fdata),
		&session.Values, s.Codecs...); err != nil {
		return err
	}
	return nil
}
