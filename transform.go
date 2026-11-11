package reqforge

import (
	stdbytes "bytes"
	"encoding/json"
	stdio "io"
	"net/http"

	"github.com/0LuigiCode0/errors"
	"github.com/google/uuid"
)

// inTransform — сериализация тела запроса в http.Request.
type inTransform[tIn any] func(v tIn, req *http.Request) error

// outTransform — десериализация тела ответа из http.Response.
type outTransform[tOut any] func(v tOut, resp *http.Response) error

// UIDPathInTransform дописывает строковое представление UUID в конец пути запроса.
func UIDPathInTransform(v uuid.UUID, req *http.Request) error {
	req.URL.Path += v.String()
	return nil
}

// StrPathInTransform дописывает строку в конец пути запроса.
func StrPathInTransform(v string, req *http.Request) error {
	req.URL.Path += v
	return nil
}

// JsonInTransform — тело запроса в JSON + Content-Type: application/json; GetBody задан для перечитывания при редиректах/ретраях.
func JsonInTransform[tIn any](v tIn, req *http.Request) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	reader := stdbytes.NewReader(body)
	req.Body = stdio.NopCloser(reader)
	req.GetBody = func() (stdio.ReadCloser, error) {
		return stdio.NopCloser(stdbytes.NewReader(body)), nil
	}
	req.ContentLength = int64(len(body))
	req.Header.Set(HeaderContentType, ContentTypeJson)

	return nil
}

// JsonOutTransform — тело ответа из JSON в v.
func JsonOutTransform[tOut any](v tOut, resp *http.Response) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// JsonEOFOutTransform — как JsonOutTransform, но пустое тело (EOF) — не ошибка, v остаётся нулевым.
func JsonEOFOutTransform[tOut any](v tOut, resp *http.Response) error {
	err := JsonOutTransform(v, resp)
	if errors.Is(err, stdio.EOF) || errors.Is(err, stdio.ErrUnexpectedEOF) {
		return nil
	}

	return err
}

// RawInTransform — сырые байты тела запроса; GetBody для перечитывания; Content-Type не трогает.
func RawInTransform[tIn ~[]byte](v tIn, req *http.Request) error {
	reader := stdbytes.NewReader(v)
	req.Body = stdio.NopCloser(reader)
	req.GetBody = func() (stdio.ReadCloser, error) {
		return stdio.NopCloser(stdbytes.NewReader(v)), nil
	}
	req.ContentLength = int64(len(v))

	return nil
}

// RawOutTransform — тело ответа целиком (сырые байты) в v; пустое тело — не ошибка (v = nil).
func RawOutTransform[tOut ~[]byte](v *tOut, resp *http.Response) error {
	buf, err := stdio.ReadAll(resp.Body)
	*v = buf
	return err
}

// RawEOFOutTransform — как RawOutTransform, но EOF/ErrUnexpectedEOF игнорируются.
func RawEOFOutTransform[tOut ~[]byte](v *tOut, resp *http.Response) error {
	buf, err := stdio.ReadAll(resp.Body)
	if errors.Is(err, stdio.EOF) || errors.Is(err, stdio.ErrUnexpectedEOF) {
		return nil
	}
	*v = buf

	return err
}
