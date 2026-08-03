package reqforge

import (
	"net/http"
	"net/url"
)

// Forge — shared HTTP client template with common configuration (headers, query params, middleware chain).
// A Forge can be reused to create multiple Requests, each cloning its state for isolation.
type Forge[tMeta any] struct {
	core[tMeta]

	host string
}

// NewForge creates a new Forge with the given base URL (host) and default http.RoundTripper.
// All Requests created from this Forge will inherit its configuration unless overridden.
func NewForge[tMeta any](host string, baseTrip http.RoundTripper) *Forge[tMeta] {
	return &Forge[tMeta]{
		host: host,
		core: core[tMeta]{
			baseTrip:  baseTrip, // сохраняется для доступа из Ctx.SetBaseTrip (например, подмена на mock)
			trip:      wrapBaseTrip[tMeta], // начальный элемент цепочки middleware — вызывает baseTrip
			header:    http.Header{},
			urlValues: url.Values{},
		},
	}
}

// SetHeader — задать HTTP-заголовки (переопределяет существующие).
func (f *Forge[tMeta]) SetHeader(header http.Header) *Forge[tMeta] {
	f.setHeader(header)
	return f
}

// SetContentType — установить заголовок Content-Type.
func (f *Forge[tMeta]) SetContentType(contentType ContentType) *Forge[tMeta] {
	f.setContentType(contentType)
	return f
}

// SetUserAgent — установить заголовок User-Agent.
func (f *Forge[tMeta]) SetUserAgent(userAgent string) *Forge[tMeta] {
	f.setUserAgent(userAgent)
	return f
}

// SetUrlValue — задать параметры URL-строки (переопределяет существующие).
func (f *Forge[tMeta]) SetUrlValue(urlValues url.Values) *Forge[tMeta] {
	f.setUrlValue(urlValues)
	return f
}

// AddTrip — добавить middleware в цепочку выполнения запросов.
// Middleware добавляются в порядке вызова, каждый обёртывает предыдущий.
func (f *Forge[tMeta]) AddTrip(chainTrips ...ChainTrip[tMeta]) *Forge[tMeta] {
	f.addTrip(chainTrips)
	return f
}

// AddHeader — добавить HTTP-заголовки к уже существующим.
func (f *Forge[tMeta]) AddHeader(header http.Header) *Forge[tMeta] {
	f.addHeader(header)
	return f
}

// AddUrlValue — добавить параметры URL-строки к уже существующим.
func (f *Forge[tMeta]) AddUrlValue(urlValues url.Values) *Forge[tMeta] {
	f.addUrlValue(urlValues)
	return f
}

// SetEncoder — задать кодировщик (serializer) для тела запроса.
func (f *Forge[tMeta]) SetEncoder(encoder encoder[any, tMeta]) *Forge[tMeta] {
	f.setEncoder(encoder)
	return f
}

// SetDecoder — задать декодер (deserializer) для тела ответа.
func (f *Forge[tMeta]) SetDecoder(decoder decoder[any, tMeta]) *Forge[tMeta] {
	f.setDecoder(decoder)
	return f
}
