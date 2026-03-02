package middleware

import (
	"errors"
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog/log"
)

var (
	ErrEmptyPrompt         = errors.New("prompt could not be empty")
	ErrInvalidMaxTokens    = errors.New("invalid number of input tokens")
	ErrInvalidTemperature  = errors.New("invalid temperature. Value must be between 0 and 1")
	ErrInternalServerError = errors.New("internal server error")
)

type ErrorResponse struct {
	Error   string `json:"error" description:"Error message"`
	Code    int    `json:"code" description:"HTTP status code"`
	Details string `json:"details" description:"Additional error details"`
}

// HandleError writes an error response
func HandleError(resp *restful.Response, err error, statusCode int) {
	log.Error().
		Err(err).
		Int("status", statusCode).
		Msg("Request error")

	errResp := ErrorResponse{
		Error:   err.Error(),
		Code:    statusCode,
		Details: "",
	}

	resp.WriteHeaderAndEntity(statusCode, errResp)
}

func RecoverPanic(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Interface("panic", r).
				Str("path", req.Request.URL.Path).
				Msg("Panic recovered")

			HandleError(resp, ErrInternalServerError, http.StatusInternalServerError)
		}
	}()

	chain.ProcessFilter(req, resp)
}
