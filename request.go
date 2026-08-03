package reqforge

import (
	"context"
	"net/http"
	"net/url"
	"slices"
)

// Request — экземпляр HTTP-запроса для конкретного endpoint (method + path).
// Создаётся из Forge и содержит изолированную копию его конфигурации.
// Generic parameters: tIn — тип тела запроса, tOut — тип тела ответа, tMeta — произвольные метаданные middleware.
type Request[tIn any, tOut, tMeta any] struct {
	core[tMeta]

	url    string
	method Method

	encoder encoder[tIn, tMeta]
	decoder decoder[*tOut, tMeta]
}

// NewRequest создаёт новый Request на основе Forge для заданного HTTP-метода и пути.
// Возвращённый Request содержит глубокую копию конфигурации Forge.
func NewRequest[tIn any, tOut, tMeta any](f *Forge[tMeta], method Method, path string) *Request[tIn, tOut, tMeta] {
	return &Request[tIn, tOut, tMeta]{
		url:    f.host + path,
		method: method,
		core:   f.core.clone(),
	}
}

// SetHeader — задать HTTP-заголовки (переопределяет существующие).
func (r *Request[tIn, tOut, tMeta]) SetHeader(header http.Header) *Request[tIn, tOut, tMeta] {
	r.setHeader(header)
	return r
}

// SetContentType — установить заголовок Content-Type.
func (r *Request[tIn, tOut, tMeta]) SetContentType(contentType ContentType) *Request[tIn, tOut, tMeta] {
	r.setContentType(contentType)
	return r
}

// SetUserAgent — установить заголовок User-Agent.
func (r *Request[tIn, tOut, tMeta]) SetUserAgent(userAgent string) *Request[tIn, tOut, tMeta] {
	r.setUserAgent(userAgent)
	return r
}

// SetUrlValue — задать параметры URL-строки (переопределяет существующие).
func (r *Request[tIn, tOut, tMeta]) SetUrlValue(urlValues url.Values) *Request[tIn, tOut, tMeta] {
	r.setUrlValue(urlValues)
	return r
}

// AddTrip — добавить middleware в цепочку выполнения этого запроса.
func (r *Request[tIn, tOut, tMeta]) AddTrip(chainTrips ...ChainTrip[tMeta]) *Request[tIn, tOut, tMeta] {
	r.addTrip(chainTrips)
	return r
}

// AddHeader — добавить HTTP-заголовки к уже существующим.
func (r *Request[tIn, tOut, tMeta]) AddHeader(header http.Header) *Request[tIn, tOut, tMeta] {
	r.addHeader(header)
	return r
}

// AddUrlValue — добавить параметры URL-строки к уже существующим.
func (r *Request[tIn, tOut, tMeta]) AddUrlValue(urlValues url.Values) *Request[tIn, tOut, tMeta] {
	r.addUrlValue(urlValues)
	return r
}

// SetEncoder — задать кодировщик для тела этого запроса (переопределяет Forge).
func (r *Request[tIn, tOut, tMeta]) SetEncoder(encoder encoder[tIn, tMeta]) *Request[tIn, tOut, tMeta] {
	r.encoder = encoder
	return r
}

// SetDecoder — задать декодер для тела этого запроса (переопределяет Forge).
func (r *Request[tIn, tOut, tMeta]) SetDecoder(decoder decoder[*tOut, tMeta]) *Request[tIn, tOut, tMeta] {
	r.decoder = decoder
	return r
}

// Do — выполнить HTTP-запрос.
// Параметры: ctx — контекст, in — тело запроса, meta — метаданные для middleware.
// Возвращает: распарсенное тело ответа и ошибку (если возникла).
func (r *Request[tIn, tOut, tMeta]) Do(ctx context.Context, in tIn, meta *tMeta) (*tOut, error) {
	// Создаём tripper — адаптер http.RoundTripper, который выполняет цепочку Trip.
	// Ctx содержит метаданные и ссылку на базовый RoundTripper (можно подменить через SetBaseTrip).
	tripper := &tripper[tMeta]{
		ctx: &Ctx[tMeta]{
			Meta:    meta,
			tripper: r.baseTrip,
		},
		trip: r.trip,
	}

	reqCtx, err := http.NewRequestWithContext(ctx, r.method, r.url, nil)
	if err != nil {
		return nil, err
	}

	// Копируем заголовки из конфигурации — чтобы не мутировать исходную core.
	reqCtx.Header = make(http.Header, len(r.header))
	for k, v := range r.header {
		reqCtx.Header[k] = slices.Clone(v)
	}
	urlValues := make(url.Values, len(r.urlValues))
	for k, v := range r.urlValues {
		urlValues[k] = slices.Clone(v)
	}
	reqCtx.URL.RawQuery = urlValues.Encode()

	// Применяем кодировщик: сначала пробуем Request-специфичный, затем Forge-по умолчанию.
	if r.encoder != nil {
		err = r.encoder(tripper.ctx, in, reqCtx)
	} else if r.core.encoder != nil {
		err = r.core.encoder(tripper.ctx, in, reqCtx)
	}
	if err != nil {
		return nil, err
	}

	c := http.Client{Transport: tripper}
	resp, err := c.Do(reqCtx)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Десериализуем тело ответа: сначала Request-специфичный decoder, затем Forge-по умолчанию.
	out := new(tOut)
	if r.decoder != nil {
		err = r.decoder(tripper.ctx, out, resp)
	} else if r.core.decoder != nil {
		err = r.core.decoder(tripper.ctx, out, resp)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}
