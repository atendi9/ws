package xrequest

import (
	"io"
	"net/http"

	"github.com/atendi9/ws/io/pkg/errors"
)

// Response wraps the standard http.Response to add utility methods.
type Response struct {
	*http.Response
}

// Ok returns true if the status code is between 200 and 299.
func (r *Response) Ok() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

func (r *Response) Err() error {
	if !r.Ok() {
		b, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		return errors.NewEngineError(r.Request.Context(), "ResponseError", string(b), nil)
	}
	return nil
}
