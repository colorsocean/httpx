package main

import (
	"errors"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/colorsocean/httpx"
	"github.com/gorilla/mux"
)

type RootResponse struct {
	SomeData string
}

func Validate(r *http.Request) {
	httpx.C(r).Error("domain1", "type1", "desc1", "targt1")
	httpx.C(r).Error("domain2", "type2", "desc2", "targt2")
	httpx.C(r).Panic(errors.New("Some shitty internal error"))
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	var resp RootResponse
	httpx.C(r).Payload(&resp)

	resp.SomeData = "point 111"

	Validate(r)

	resp.SomeData = "point 222"
}

func OtherHandler(w http.ResponseWriter, r *http.Request) {
	var resp RootResponse
	httpx.C(r).Payload(&resp)

	resp.SomeData = "point 111"

	Validate(r)

	resp.SomeData = "point 222"
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/path", RootHandler)
	r.HandleFunc("/oth", OtherHandler)

	n := negroni.New()
	n.Use(&httpx.JsonResponseMiddleware{Debug: true})
	n.Use(&OtherMiddleware{})
	n.UseHandler(r)
	n.Run(":8080")
}

type OtherMiddleware struct {
}

func (this *OtherMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(w, r)
}
