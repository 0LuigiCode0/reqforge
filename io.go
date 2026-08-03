package reqforge

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type (
	encoder[tValue, tMeta any] func(ctx *Ctx[tMeta], v tValue, req *http.Request) error
	decoder[tValue, tMeta any] func(ctx *Ctx[tMeta], v tValue, resp *http.Response) error
)

func EmptyEncode[t, tMeta any](*Ctx[tMeta], t, *http.Request) error  { return nil }
func EmptyDecode[t, tMeta any](*Ctx[tMeta], t, *http.Response) error { return nil }

func JsonEncode[t, tMeta any](c *Ctx[tMeta], v t, httpReq *http.Request) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)
	httpReq.Body = io.NopCloser(reader)
	httpReq.ContentLength = int64(len(body))
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	httpReq.Header.Set(string(HeaderContentType), string(ContentTypeJson))
	return nil
}

func JsonDecode[t, tMeta any](c *Ctx[tMeta], v t, resp *http.Response) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
