package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Responser interface {
	Response(http.ResponseWriter, *interface{}) error
}

type (
	Accept struct {
		ContentType string
		Marshaler   func(interface{}) ([]byte, error)
	}
	ContentType string
	Error       struct {
		Error      error
		StatusCode int
		Accept     Accept
	}
	Header     http.Header
	StatusCode int
)

func (a Accept) Response(w http.ResponseWriter, v *interface{}) error {
	if a.ContentType != "" {
		if err := ContentType(a.ContentType).Response(w, nil); err != nil {
			return errors.WithStack(err)
		}
	}

	if a.Marshaler != nil {
		b, err := a.Marshaler(*v)
		if err != nil {
			return errors.WithStack(err)
		}

		*v = string(b)
	}

	return nil
}

func (c ContentType) Response(w http.ResponseWriter, _ *interface{}) error {
	w.Header().Set("Content-Type", string(c))
	return nil
}

func (e Error) Response(w http.ResponseWriter, v *interface{}) error {
	if e.Error == nil {
		return nil
	}

	w.WriteHeader(e.StatusCode)
	if e.Accept.Marshaler != nil {
		nv := interface{}(e.Error)
		*v = nv
		return e.Accept.Response(w, &nv)
	}

	return errors.WithStack(e.Error)
}

func (h Header) Response(w http.ResponseWriter, _ *interface{}) error {
	for k, v := range h {
		w.Header().Del(k)
		for _, v := range v {
			w.Header().Add(k, v)
		}
	}

	return nil
}

func (s StatusCode) Response(w http.ResponseWriter, _ *interface{}) error {
	w.WriteHeader(int(s))
	return nil
}

func (s StatusCode) StatusCode() int {
	return int(s)
}

func Acceptable(r *http.Request, accepts ...Accept) Responser {
	accept := r.Header.Get("Accept")
	if (accept == "" || accept == "*/*") && len(accepts) > 0 {
		return accepts[0]
	}

	for _, a := range accepts {
		if strings.Contains(accept, a.ContentType) {
			return a
		}
	}

	err := Error{
		Error:      errors.New("unable to find an acceptable media type"),
		StatusCode: http.StatusUnsupportedMediaType,
	}
	if len(accepts) > 0 {
		err.Accept = accepts[0]
	}

	return err
}

func Write(w http.ResponseWriter, v interface{}, responses ...Responser) error {
	type statusCoder interface {
		StatusCode() int
	}

	var statusCode StatusCode
	for _, r := range responses {
		switch t := r.(type) {
		case StatusCode:
			statusCode = t
		case statusCoder:
			statusCode = StatusCode(t.StatusCode())
		default:
			if err := r.Response(w, &v); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	// defer writing the status code until after everything else since an
	// Error.Response, for example, may write a status code
	if statusCode != (StatusCode(0)) {
		statusCode.Response(w, nil)
	}

	_, err := fmt.Fprint(w, v)
	return errors.WithStack(err)
}
