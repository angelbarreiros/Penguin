package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	routerErrors "github.com/angelbarreiros/Penguin/router/errors"
)

// DeserializeBodyWithLimit decodes a single JSON value of at most maxBytes
// into dto. Empty bodies return ErrRequestBodyMissing; trailing data after the
// JSON value is rejected.
func DeserializeBodyWithLimit(r *http.Request, dto any, maxBytes int64) error {
	if r == nil || r.Body == nil || r.Body == http.NoBody {
		return routerErrors.ErrRequestBodyMissing()
	}
	var body io.ReadCloser = http.MaxBytesReader(nil, r.Body, maxBytes)
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return routerErrors.ErrRequestBodyTooLarge(int(maxBytes))
		}
		return routerErrors.ErrRequestBodyInvalid("could not read body")
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return routerErrors.ErrRequestBodyMissing()
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if err = decoder.Decode(dto); err != nil {
		return routerErrors.ErrRequestBodyInvalid(describeJSONError(err))
	}
	if decoder.More() {
		return routerErrors.ErrRequestBodyInvalid("unexpected data after JSON value")
	}

	return nil
}

func describeJSONError(err error) string {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxErr):
		return "malformed JSON"
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return "field '" + typeErr.Field + "' has the wrong type"
		}
		return "wrong JSON type"
	case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, io.EOF):
		return "unexpected end of JSON"
	default:
		return "malformed JSON"
	}
}
