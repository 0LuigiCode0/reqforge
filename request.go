package reqforge

import (
	stdcontext "context"
	"net/http"
	"net/url"
	stdpath "path"
	stdslices "slices"

	"github.com/0LuigiCode0/errors"
)

// Request — запрос к конкретному endpoint (method + path); изолированная копия конфигурации Forge.
type Request[tIn, tOut, tMeta any] struct {
	core[tMeta]

	url    string
	method Method

	// inTransform — сериализация тела запроса (опционально).
	inTransform inTransform[tIn]
	// outTransform — десериализация тела ответа (опционально).
	outTransform outTransform[*tOut]
}

// NewRequest — Request на основе Forge для заданного HTTP-метода и пути.
func NewRequest[tIn, tOut, tMeta any](f *Forge[tMeta], method Method, path string) *Request[tIn, tOut, tMeta] {
	uri := stdpath.Join(f.host, path)
	if f.schema == "" {
		uri = f.host + path
	}

	return &Request[tIn, tOut, tMeta]{
		url:    f.schema + uri,
		method: method,
		core:   f.clone(),
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
func (r *Request[tIn, tOut, tMeta]) AddTrip(chainTrips ...ChainTrip) *Request[tIn, tOut, tMeta] {
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

// SetInTransform — трансформатор входа (сериализация тела запроса) для этого запроса.
func (r *Request[tIn, tOut, tMeta]) SetInTransform(t inTransform[tIn]) *Request[tIn, tOut, tMeta] {
	r.inTransform = t
	return r
}

// SetOutTransform — трансформатор выхода (десериализация тела ответа) для этого запроса.
func (r *Request[tIn, tOut, tMeta]) SetOutTransform(t outTransform[*tOut]) *Request[tIn, tOut, tMeta] {
	r.outTransform = t
	return r
}

// Do — выполняет HTTP-запрос через middleware-цепочку, с inTransform и outTransform.
func (r *Request[tIn, tOut, tMeta]) Do(ctx stdcontext.Context, in tIn, meta *Meta[tMeta]) (*tOut, error) {
	if meta.BaseTrip == nil {
		meta.BaseTrip = r.baseTrip
	}
	ctx = meta.WithCtx(ctx)

	req, err := http.NewRequestWithContext(ctx, r.method, r.url, http.NoBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// копия, чтобы не мутировать core запроса
	req.Header = make(http.Header, len(r.header))
	for k, v := range r.header {
		req.Header[k] = stdslices.Clone(v)
	}
	urlValues := make(url.Values, len(r.urlValues))
	for k, v := range r.urlValues {
		urlValues[k] = stdslices.Clone(v)
	}
	req.URL.RawQuery = urlValues.Encode()

	// трансформатор входа (если задан)
	if r.inTransform != nil {
		if err = r.inTransform(in, req); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	tripper := &tripper[tMeta]{
		meta: meta,
		trip: r.trip,
	}
	c := http.Client{Transport: tripper}
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	// трансформатор выхода (если задан)
	var out *tOut
	if resp.Body != nil && r.outTransform != nil {
		out = new(tOut)
		if err = r.outTransform(out, resp); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return out, nil
}
