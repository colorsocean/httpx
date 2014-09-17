package main

// Global error reporting:
// Logrus Hook -> Report errors
// Logrus trace hook
// Logrus formatter?

// JsonResponse middleware (with panic callback?)
// + JRMeta service
// + Return payloads?

// Integrate above into auth
// Change Mailer

// Identity

// MEET MOSCOW 14:00

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/colorsocean/httpx"
	"github.com/go-martini/martini"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	m := martini.Classic()

	m.Handlers(
		httpx.ProfilerMMW(),
		httpx.SessionCookieMMW(httpx.SessionCookieOptions{
			AuthLifetime:  6 * time.Second,
			VisitLifetime: 3 * time.Second,

			Name:     "session_cookie",
			Domain:   "example.com",
			Secure:   false,
			HashKey:  []byte("340ef7baa43c4d6eac4128524cdc5016"),
			BlockKey: []byte("bde17e7c3ead4643a95cb3736c0332ad"),
		}),
		martini.Recovery(),
	)

	m.Get("/", func(w http.ResponseWriter, sc *httpx.SessionCookie) string {
		w.Header().Set("content-type", "text/html")
		body := fmt.Sprintln("<pre>Hello, kitty!\nAuth:", sc.AuthToken(), "\nVisi:", sc.VisitToken())
		panic(body)
		return body
	})

	type ResponsePayload struct {
		AuthToken  string
		VisitToken string
	}

	m.Group("/api", func(r martini.Router) {
		r.Group("/1", func(r martini.Router) {
			r.Get("/ep1", func(sc *httpx.SessionCookie, jr *httpx.JsonResponse) {
				pl := new(ResponsePayload)
				jr.Payload(pl)

				pl.AuthToken = sc.AuthToken()
				pl.VisitToken = sc.VisitToken()

				//panic("awful-error")
				//var a = 0
				//var b = 2 / a
				//fmt.Print(b)
			})
		})
	}, httpx.JsonResponseMMW(httpx.JsonResponseOptions{Debug: true}))

	logrus.Fatalln(http.ListenAndServe(":8080", m))
}
