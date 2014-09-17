package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
)

/****************************************************************
** JsonResponse
********/

type JsonResponseError struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Desc   string `json:"desc"`
	Target string `json:"target"`
	Trace  string `json:"trace,omitempty"`
}

type JsonResponseWarn struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Desc   string `json:"desc"`
	Target string `json:"target"`
}

type JsonResponseMeta struct {
	Ise    bool                `json:"ise"`
	Errors []JsonResponseError `json:"errors"`
	Warns  []JsonResponseWarn  `json:"warns"`
}

type JsonResponseData struct {
	Meta    JsonResponseMeta `json:"meta"`
	Payload interface{}      `json:"payload"`
}

/****************************************************************
** Martini middleware
********/

type JsonResponse struct {
	data        *JsonResponseData
	options     JsonResponseOptions
	w           http.ResponseWriter
	alreadySent bool
}

func (this JsonResponse) log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"module": "json_response",
	})
}

func (this *JsonResponse) Error(domain, type_, desc, target string) {
	this.error(domain, type_, desc, target, "")
}

func (this *JsonResponse) error(domain, type_, desc, target, trace string) {
	this.data.Meta.Errors = append(this.data.Meta.Errors, JsonResponseError{
		Domain: domain,
		Type:   type_,
		Desc:   desc,
		Target: target,
		Trace:  trace,
	})
}

func (this *JsonResponse) Ise(err interface{}) {
	trace := ""
	if this.options.Debug {
		traceBytes := make([]byte, 32*1024)
		count := runtime.Stack(traceBytes, true)
		traceBytes = traceBytes[0:count]
		trace = string(traceBytes)
	}
	this.error(DomainGeneric, ErrTypeISE, fmt.Sprint(err), "", trace)

	this.data.Meta.Ise = true
}

func (this *JsonResponse) Warn(domain, type_, desc, target string) {
	this.data.Meta.Warns = append(this.data.Meta.Warns, JsonResponseWarn{
		Domain: domain,
		Type:   type_,
		Desc:   desc,
		Target: target,
	})
}

func (this *JsonResponse) Payload(payload interface{}) {
	this.data.Payload = payload
}

func (this *JsonResponse) Send(status0 ...int) error {
	if this.alreadySent {
		return ErrAlreadySent
	}

	this.alreadySent = true

	status := http.StatusOK
	if this.data.Meta.Ise {
		status = http.StatusInternalServerError
	} else if len(this.data.Meta.Errors) > 0 {
		if len(status0) > 0 && status0[0] >= 400 {
			status = status0[0]
		} else {
			status = http.StatusInternalServerError
		}
	} else if len(status0) > 0 {
		status = status0[0]
	}

	sendJson := func(data []byte) error {
		this.w.Header().Set(headerContentType, headerContentTypeApplicationJson)
		this.w.WriteHeader(status)
		_, err := bytes.NewBuffer(data).WriteTo(this.w)
		return err
	}

	data, err := this.getDataJson()
	if err != nil {
		this.log().Errorln(err)

		status = http.StatusInternalServerError
		this.data.Payload = nil
		this.Ise(ErrCanNotSerializePayload.Error())

		data, err2 := this.getDataJson()
		if err2 != nil {
			this.log().Errorln(err2)

			this.w.Header().Set(headerContentType, headerContentTypeApplicationJson)
			this.w.WriteHeader(http.StatusInternalServerError)
			_, err3 := fmt.Fprintln(this.w, `{"meta":{"ise":true}}`)

			if err3 != nil {
				this.log().Errorln(err3)
			}

			return err2
		}

		err2 = sendJson(data)
		if err2 != nil {
			this.log().Errorln(err2)
		}
		return err
	}

	return sendJson(data)
}

func (this *JsonResponse) getDataJson() ([]byte, error) {
	body, err := json.MarshalIndent(this.data, "", " ")
	return body, err
}

type JsonResponseOptions struct {
	Debug bool
}

func JsonResponseMMW(ops JsonResponseOptions) martini.Handler {
	return func(w http.ResponseWriter, r *http.Request, c martini.Context) {
		jr := &JsonResponse{
			options:     ops,
			data:        &JsonResponseData{},
			w:           w,
			alreadySent: false,
		}

		defer func() {
			if rec := recover(); rec != nil {
				jr.Ise(rec)
			}

			err := jr.Send()
			if err != nil {
				jr.log().Errorln(err)
			}
		}()

		whook := &jsonResponseResponseWriterHook{w}

		c.Map(jr)

		c.MapTo(whook, (*http.ResponseWriter)(nil))

		c.Next()
	}
}

type jsonResponseResponseWriterHook struct {
	w http.ResponseWriter
}

func (this *jsonResponseResponseWriterHook) Header() http.Header {
	return this.w.Header()
}

func (this *jsonResponseResponseWriterHook) Write(data []byte) (int, error) {
	return 0, ErrOnlyAutomaticResponse
}

func (this *jsonResponseResponseWriterHook) WriteHeader(status int) {
	this.w.WriteHeader(status)
}
