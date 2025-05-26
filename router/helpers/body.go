package helpers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	routerErrors "github.com/angelbarreiros/Penguin/router/errors"
)

func DeserializeBodyWithLimit(r *http.Request, dto any, maxBytes int64) error {
	var body io.ReadCloser = http.MaxBytesReader(nil, r.Body, maxBytes)
	defer body.Close()

	bytes, err := io.ReadAll(body)
	if err != nil {
		if err == io.EOF {
			return routerErrors.ErrRequestBodyMissing()
		}
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return routerErrors.ErrRequestBodyTooLarge(int(maxBytes))
		}
		return routerErrors.ErrRequestBodyInvalid(err.Error())
	}

	if err = json.Unmarshal(bytes, dto); err != nil {
		return err
	}

	return nil
}
