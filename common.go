package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

const (
	headerContentType   = "Content-Type"
	headerContentLength = "Content-Length"

	headerContentTypeTextHtml        = "text/html; encoding=utf-8"
	headerContentTypeApplicationJson = "application/json; charset=utf-8"
	headerContentTypeTextJavascript  = "text/javascript; charset=utf-8"

	DomainGeneric = "generic"

	ErrTypeISE = "ise"
)

var (
	ErrOnlyAutomaticResponse  = errors.New("Please, don't write anything to ResponseWriter directly")
	ErrCanNotSerializePayload = errors.New("Can not serialize payload")
	ErrAlreadySent            = errors.New("Response already sent")
)

/****************************************************************
** Common Helpers
********/

func ListenAndServeUnix(sock string, handler http.Handler) error {
	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	} else {
		err := http.Serve(l, handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func HtmlPrintf(w http.ResponseWriter, code int, format string, a ...interface{}) error {
	w.Header().Set(headerContentType, headerContentTypeTextHtml)
	w.WriteHeader(code)
	_, err := fmt.Fprintf(w, format, a...)
	return err
}

func IsXhr(req *http.Request) bool {
	return req.Header.Get("HTTP_X_REQUESTED_WITH") == "XMLHttpRequest"
}

func WriteJsonBody(w http.ResponseWriter, obj interface{}, status ...int) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	s := http.StatusOK
	if len(status) > 0 {
		s = status[0]
	}

	w.Header().Set(headerContentType, headerContentTypeApplicationJson)
	w.Header().Set(headerContentLength, fmt.Sprint(len(data)))
	w.WriteHeader(s)
	_, err = w.Write(data)

	if err == nil {
		return err
	}

	return nil
}

func ReadJsonBody(r *http.Request, outObj interface{}) error {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, outObj)

	if err != nil {
		return err
	}

	return nil
}

/****************************************************************
** Kawaii HTTP Responses
********/

func kawaii(w http.ResponseWriter, status int, smiley, heading, format string, args ...interface{}) error {
	kawaii_tpl := `<pre><h1>%s %d %s</h1><br/>%s<br/><br/>%s<br/><br/></pre>`
	return HtmlPrintf(w, status, kawaii_tpl, smiley, status, smiley, heading, fmt.Sprintf(format, args...))
}

func Kawaii500(w http.ResponseWriter, format string, args ...interface{}) error {
	return kawaii(w, http.StatusInternalServerError, "(_O_)", "Internal Server Error!", format, args...)
}

func Kawaii404(w http.ResponseWriter, format string, args ...interface{}) error {
	return kawaii(w, http.StatusNotFound, "=(^_^)=", "Not Found!", format, args...)
}

func Kawaii403(w http.ResponseWriter, format string, args ...interface{}) error {
	return kawaii(w, http.StatusForbidden, "(_X_)", "Forbidden!", format, args...)
}

func Kawaii401(w http.ResponseWriter, format string, args ...interface{}) error {
	return kawaii(w, http.StatusUnauthorized, "(_x_)", "Not Authorized!", format, args...)
}
