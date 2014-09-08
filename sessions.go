package sessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

type Store interface {
	sessions.Store
}

type Options struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

type Session interface {
	Get(key interface{}) interface{}
	Set(key interface{}, val interface{})
	Delete(key interface{})
	Clear()
	AddFlash(value interface{}, vars ...string)
	Flashes(vars ...string) []interface{}
	Save() error
	Options(Options)
}

func Sessions(name string, store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := &session{
			name:    name,
			request: c.Request,
			writer:  c.Writer,
			store:   store,
		}

		c.Set("session", s)
		defer context.Clear(c.Request)
		c.Next()
	}
}

type session struct {
	name    string
	request *http.Request
	writer  http.ResponseWriter
	store   Store
	session *sessions.Session
}

func (s *session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

func (s *session) Set(key interface{}, value interface{}) {
	s.Session().Values[key] = value
}

func (s *session) Delete(key interface{}) {
	delete(s.Session().Values, key)
}

func (s *session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

func (s *session) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
}

func (s *session) Flashes(vars ...string) []interface{} {
	return s.Session().Flashes(vars...)
}

func (s *session) Save() error {
	return s.Session().Save(s.request, s.writer)
}

func (s *session) Session() *sessions.Session {
	if s.session == nil {
		if session, err := s.store.Get(s.request, s.name); err != nil {
			panic(err)
		} else {
			s.session = session
		}
	}

	return s.session
}

func (s *session) Options(options Options) {
	s.Session().Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}
