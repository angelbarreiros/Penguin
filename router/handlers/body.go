package handlers

import (
	routerErrors "angelotero/commonBackend/router/errors"
	"errors"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
)

func DeserializeBodyWithLimit(r *http.Request, dto any, maxBytes int64) error {
	body := http.MaxBytesReader(nil, r.Body, maxBytes)
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

	if err = sonic.Unmarshal(bytes, dto); err != nil {
		return err
	}

	return nil
}
