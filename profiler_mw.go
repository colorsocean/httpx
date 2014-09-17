package httpx

import (
	"net/http"
	"time"

	"github.com/go-martini/martini"
)

func ProfilerMMW() martini.Handler {
	return func(w http.ResponseWriter, r *http.Request, c martini.Context) {

		phook := &profilerResponseWriterHook{time.Now(), w, false}

		c.MapTo(phook, (*http.ResponseWriter)(nil))

		c.Next()
	}
}

type profilerResponseWriterHook struct {
	startTime time.Time
	w         http.ResponseWriter
	isSet     bool
}

func (this *profilerResponseWriterHook) ensureHeaderSet() {
	if !this.isSet {
		this.w.Header().Set("Debug-Request-Time", time.Now().Sub(this.startTime).String())
		this.isSet = true
	}
}

func (this *profilerResponseWriterHook) Header() http.Header {
	return this.w.Header()
}

func (this *profilerResponseWriterHook) Write(b []byte) (int, error) {
	this.ensureHeaderSet()
	return this.w.Write(b)
}

func (this *profilerResponseWriterHook) WriteHeader(s int) {
	this.ensureHeaderSet()
	this.w.WriteHeader(s)
}
