package sessions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type requestOptions struct {
	Method  string
	URL     string
	Headers map[string]string
}

func request(server *gin.Engine, options requestOptions) *httptest.ResponseRecorder {
	if options.Method == "" {
		options.Method = "GET"
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(options.Method, options.URL, nil)

	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	server.ServeHTTP(w, req)

	if err != nil {
		panic(err)
	}

	return w
}

func newServer() *gin.Engine {
	g := gin.New()
	store := NewCookieStore([]byte("secret123"))
	g.Use(Middleware("my_session", store))

	return g
}

func sessionContext(fn func(session Session)) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := Get(c)
		fn(session)
	}
}

func TestSessions(t *testing.T) {
	g := newServer()

	g.GET("/testsession", sessionContext(func(session Session) {
		session.Set("hello", "world")
		session.Save()
	}))

	g.GET("/show", sessionContext(func(session Session) {
		if session.Get("hello") != "world" {
			t.Error("Session writing failed")
		}
	}))

	r1 := request(g, requestOptions{URL: "/testsession"})
	request(g, requestOptions{
		URL: "/show",
		Headers: map[string]string{
			"Cookie": r1.Header().Get("Set-Cookie"),
		},
	})
}

func TestSessionsDelete(t *testing.T) {
	g := newServer()

	g.GET("/testsession", sessionContext(func(session Session) {
		session.Set("hello", "world")
		session.Delete("hello")
		session.Save()
	}))

	g.GET("/show", sessionContext(func(session Session) {
		if session.Get("hello") == "world" {
			t.Error("Session delete failed")
		}
	}))

	r1 := request(g, requestOptions{URL: "/testsession"})
	request(g, requestOptions{
		URL: "/show",
		Headers: map[string]string{
			"Cookie": r1.Header().Get("Set-Cookie"),
		},
	})
}

func TestOptions(t *testing.T) {
	g := gin.New()
	store := NewCookieStore([]byte("secret123"))
	store.Options(Options{
		Domain: "maji.moe",
	})
	g.Use(Middleware("my_session", store))

	g.GET("/", sessionContext(func(session Session) {
		session.Set("hello", "world")
		session.Options(Options{
			Path: "/foo/bar/bat",
		})
		session.Save()
	}))

	g.GET("/foo", sessionContext(func(session Session) {
		session.Set("hello", "world")
		session.Save()
	}))

	r1 := request(g, requestOptions{URL: "/"})
	r2 := request(g, requestOptions{URL: "/foo"})

	s := strings.Split(r1.Header().Get("Set-Cookie"), ";")
	if s[1] != " Path=/foo/bar/bat" {
		t.Error("Error writing path with options:", s[1])
	}

	s = strings.Split(r2.Header().Get("Set-Cookie"), ";")
	if s[1] != " Domain=maji.moe" {
		t.Error("Error writing domain with options:", s[1])
	}
}

func TestFlashes(t *testing.T) {
	g := newServer()

	g.GET("/set", sessionContext(func(session Session) {
		session.AddFlash("hello wrold")
		session.Save()
	}))

	g.GET("/show", sessionContext(func(session Session) {
		l := len(session.Flashes())

		if l != 1 {
			t.Error("Flashes count does not equal 1. Equals ", l)
		}
	}))

	g.GET("/showagain", sessionContext(func(session Session) {
		l := len(session.Flashes())
		if l != 0 {
			t.Error("Flashes count is not 0 after reading. Equals ", l)
		}
	}))

	r1 := request(g, requestOptions{URL: "/set"})
	r2 := request(g, requestOptions{
		URL: "/show",
		Headers: map[string]string{
			"Cookie": r1.Header().Get("Set-Cookie"),
		},
	})
	request(g, requestOptions{
		URL: "/showagain",
		Headers: map[string]string{
			"Cookie": r2.Header().Get("Set-Cookie"),
		},
	})
}

func TestSessionClear(t *testing.T) {
	g := newServer()
	data := map[string]string{
		"hello":  "world",
		"foo":    "bar",
		"apples": "oranges",
	}

	g.GET("/testsession", sessionContext(func(session Session) {
		for key, value := range data {
			session.Set(key, value)
		}
		session.Clear()
		session.Save()
	}))

	g.GET("/show", sessionContext(func(session Session) {
		for key, value := range data {
			if session.Get(key) == value {
				t.Error("Session clear failed")
			}
		}
	}))

	r1 := request(g, requestOptions{URL: "/testsession"})
	request(g, requestOptions{
		URL: "/show",
		Headers: map[string]string{
			"Cookie": r1.Header().Get("Set-Cookie"),
		},
	})
}
