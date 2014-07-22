package sessions

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/supershabam/vengo/github.com/gorilla/context"
)

const flashesKey = "_flash"

type Options struct {
	Path	string
	Domain	string

	MaxAge		int
	Secure		bool
	HttpOnly	bool
}

func NewSession(store Store, name string) *Session {
	return &Session{
		Values:	make(map[interface{}]interface{}),
		store:	store,
		name:	name,
	}
}

type Session struct {
	ID	string
	Values	map[interface{}]interface{}
	Options	*Options
	IsNew	bool
	store	Store
	name	string
}

func (s *Session) Flashes(vars ...string) []interface{} {
	var flashes []interface{}
	key := flashesKey
	if len(vars) > 0 {
		key = vars[0]
	}
	if v, ok := s.Values[key]; ok {

		delete(s.Values, key)
		flashes = v.([]interface{})
	}
	return flashes
}

func (s *Session) AddFlash(value interface{}, vars ...string) {
	key := flashesKey
	if len(vars) > 0 {
		key = vars[0]
	}
	var flashes []interface{}
	if v, ok := s.Values[key]; ok {
		flashes = v.([]interface{})
	}
	s.Values[key] = append(flashes, value)
}

func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.store.Save(r, w, s)
}

func (s *Session) Name() string {
	return s.name
}

func (s *Session) Store() Store {
	return s.store
}

type sessionInfo struct {
	s	*Session
	e	error
}

type contextKey int

const registryKey contextKey = 0

func GetRegistry(r *http.Request) *Registry {
	registry := context.Get(r, registryKey)
	if registry != nil {
		return registry.(*Registry)
	}
	newRegistry := &Registry{
		request:	r,
		sessions:	make(map[string]sessionInfo),
	}
	context.Set(r, registryKey, newRegistry)
	return newRegistry
}

type Registry struct {
	request		*http.Request
	sessions	map[string]sessionInfo
}

func (s *Registry) Get(store Store, name string) (session *Session, err error) {
	if info, ok := s.sessions[name]; ok {
		session, err = info.s, info.e
	} else {
		session, err = store.New(s.request, name)
		session.name = name
		s.sessions[name] = sessionInfo{s: session, e: err}
	}
	session.store = store
	return
}

func (s *Registry) Save(w http.ResponseWriter) error {
	var errMulti MultiError
	for name, info := range s.sessions {
		session := info.s
		if session.store == nil {
			errMulti = append(errMulti, fmt.Errorf(
				"sessions: missing store for session %q", name))
		} else if err := session.store.Save(s.request, w, session); err != nil {
			errMulti = append(errMulti, fmt.Errorf(
				"sessions: error saving session %q -- %v", name, err))
		}
	}
	if errMulti != nil {
		return errMulti
	}
	return nil
}

func init() {
	gob.Register([]interface{}{})
}

func Save(r *http.Request, w http.ResponseWriter) error {
	return GetRegistry(r).Save(w)
}

func NewCookie(name, value string, options *Options) *http.Cookie {
	cookie := &http.Cookie{
		Name:		name,
		Value:		value,
		Path:		options.Path,
		Domain:		options.Domain,
		MaxAge:		options.MaxAge,
		Secure:		options.Secure,
		HttpOnly:	options.HttpOnly,
	}
	if options.MaxAge > 0 {
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {

		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}

type MultiError []error

func (m MultiError) Error() string {
	s, n := "", 0
	for _, e := range m {
		if e != nil {
			if n == 0 {
				s = e.Error()
			}
			n++
		}
	}
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", s, n-1)
}
