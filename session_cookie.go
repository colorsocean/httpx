package httpx

import (
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/colorsocean/utils"
	"github.com/go-martini/martini"
	"github.com/gorilla/securecookie"
)

/****************************************************************
** Martini middleware
********/

func SessionCookieMMW(ops SessionCookieOptions) martini.Handler {
	return func(w http.ResponseWriter, r *http.Request, c martini.Context) {

		sc := &SessionCookie{
			options: ops,
		}

		sc.log().Debugln("SessionCookieMMW called")

		sc.Read(r)

		whook := &sessionCookieResponseWriterHook{sc, w, false}

		c.Map(sc)

		c.MapTo(whook, (*http.ResponseWriter)(nil))

		c.Next()
	}
}

type sessionCookieResponseWriterHook struct {
	sc    *SessionCookie
	w     http.ResponseWriter
	isSet bool
}

func (this *sessionCookieResponseWriterHook) ensureSet() {
	if !this.isSet {
		this.sc.Write(this.w)
		this.isSet = true
	}
}

func (this *sessionCookieResponseWriterHook) Header() http.Header {
	return this.w.Header()
}

func (this *sessionCookieResponseWriterHook) Write(data []byte) (int, error) {
	this.ensureSet()
	return this.w.Write(data)
}

func (this *sessionCookieResponseWriterHook) WriteHeader(status int) {
	this.ensureSet()
	this.w.WriteHeader(status)
}

/****************************************************************
** Session Cookie impl
********/

type SessionCookie struct {
	enc *securecookie.SecureCookie

	data *sessionCookieData

	options SessionCookieOptions

	once sync.Once
}

type SessionCookieOptions struct {
	AuthLifetime  time.Duration
	VisitLifetime time.Duration

	Name     string
	Domain   string
	Secure   bool
	HashKey  []byte
	BlockKey []byte
}

type sessionCookieData struct {
	AuthToken        string
	AuthTokenCreated time.Time
	AuthTokenRenewed time.Time

	VisitToken        string
	VisitTokenCreated time.Time
	VisitTokenRenewed time.Time
}

func (this SessionCookie) log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"module": "session_cookie",
	})
}

func (this *SessionCookie) init() {
	this.once.Do(func() {
		this.enc = securecookie.New(this.options.HashKey, this.options.BlockKey)

		this.data = &sessionCookieData{}

		this.check()
	})
}

func (this *SessionCookie) check() {
	now := time.Now().UTC()

	if !utils.MatchUUIDNoDashes.MatchString(this.data.AuthToken) ||
		now.After(this.data.AuthTokenRenewed.Add(this.options.AuthLifetime)) {
		this.data.AuthToken = utils.RandomUUIDNoDashes()
		this.data.AuthTokenCreated = now
	}

	if !utils.MatchUUIDNoDashes.MatchString(this.data.VisitToken) ||
		now.After(this.data.VisitTokenRenewed.Add(this.options.VisitLifetime)) {
		this.data.VisitToken = utils.RandomUUIDNoDashes()
		this.data.VisitTokenCreated = now
	}
}

func (this *SessionCookie) Reset() {
	this.init()

	now := time.Now().UTC()

	this.data.AuthToken = utils.RandomUUIDNoDashes()
	this.data.AuthTokenCreated = now

	this.data.VisitToken = utils.RandomUUIDNoDashes()
	this.data.VisitTokenCreated = now

	this.data.AuthTokenRenewed = now
	this.data.VisitTokenRenewed = now
}

func (this *SessionCookie) AuthToken() string {
	this.init()

	return this.data.AuthToken
}

func (this *SessionCookie) VisitToken() string {
	this.init()

	return this.data.VisitToken
}

func (this *SessionCookie) Read(r *http.Request) {
	this.init()

	cookie, err := r.Cookie(this.options.Name)
	if err == nil {
		err = this.enc.Decode(this.options.Name, cookie.Value, this.data)
		if err != nil {
			this.log().Warnln("Can not decode cookie")
		}
	} else {
		this.log().Debugln("No cookie present")
	}

	this.check()

	this.data.AuthTokenRenewed = time.Now().UTC()
	this.data.VisitTokenRenewed = time.Now().UTC()
}

func (this *SessionCookie) Write(w http.ResponseWriter) {
	this.init()

	this.check()

	if encoded, err := this.enc.Encode(this.options.Name, this.data); err == nil {
		age := utils.MaxDuration(this.options.AuthLifetime, this.options.VisitLifetime)

		cookie := &http.Cookie{
			Name:     this.options.Name,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Secure:   this.options.Secure,
			Domain:   this.options.Domain,
			Expires:  time.Now().UTC().Add(age),
		}

		http.SetCookie(w, cookie)
	} else {
		this.log().Errorln("Error encoding cookie", err.Error())
	}
}
