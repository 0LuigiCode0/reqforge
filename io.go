package reqforge

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// encoder — функция для сериализации тела запроса в http.Request.
// tValue — тип тела запроса, tMeta — тип метаданных из контекста.
type encoder[tValue, tMeta any] func(ctx *Ctx[tMeta], v tValue, req *http.Request) error

// decoder — функция для десериализации тела ответа из http.Response.
// tValue — указатель на тип тела ответа, tMeta — тип метаданных из контекста.
type decoder[tValue, tMeta any] func(ctx *Ctx[tMeta], v tValue, resp *http.Response) error

// EmptyEncode — пустой кодировщик: ничего не делает, возвращает nil.
func EmptyEncode[t, tMeta any](*Ctx[tMeta], t, *http.Request) error  { return nil }

// EmptyDecode — пустой декодер: ничего не делает, возвращает nil.
func EmptyDecode[t, tMeta any](*Ctx[tMeta], t, *http.Response) error { return nil }

// JsonEncode — сериализовать тело запроса в JSON и установить его в http.Request.
// GetBody устанавливается для повторного чтения тела (http.Client использует его при редиректах).
func JsonEncode[t, tMeta any](c *Ctx[tMeta], v t, httpReq *http.Request) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)
	httpReq.Body = io.NopCloser(reader)
	httpReq.ContentLength = int64(len(body)) // обязательно для корректных HTTP-запросов
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	httpReq.Header.Set(string(HeaderContentType), string(ContentTypeJson))
	return nil
}

// JsonDecode — десериализовать тело ответа из JSON в указатель на v.
func JsonDecode[t, tMeta any](c *Ctx[tMeta], v t, resp *http.Response) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
