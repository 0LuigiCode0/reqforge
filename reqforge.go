// Package reqforge — generic-билдер HTTP-запросов с middleware-цепочкой.
//
// Forge — переиспользуемый шаблон; каждый Request получает изолированную копию состояния.
//
//	forge := reqforge.NewForge[*reqforge.BaseMeta](host, baseTrip)
//	forge.SetContentType(reqforge.ContentTypeJson)
//	req := reqforge.NewRequest[RequestBody, ResponseBody, *reqforge.BaseMeta](forge, "POST", "/api/users")
//	resp, err := req.Do(ctx, requestBody, meta)
package reqforge

import (
	"net/http"
	"net/url"
)

// Forge — шаблон HTTP-клиента (host + общая конфигурация); tMeta — тип метаданных для middleware.
type Forge[tMeta any] struct {
	core[tMeta]

	// host — базовый URL (хост) для всех запросов из этого Forge.
	host   string
	schema Schema
}

// NewForge — Forge с базовым URL (host) и базовым RoundTripper, вызываемым в конце цепочки middleware.
func NewForge[tMeta any](schema Schema, host string, baseTrip Trip) *Forge[tMeta] {
	return &Forge[tMeta]{
		schema: schema,
		host:   host,
		core: core[tMeta]{
			baseTrip:  baseTrip,
			trip:      wrapBaseTrip[tMeta],
			header:    http.Header{},
			urlValues: url.Values{},
		},
	}
}

func (f *Forge[tMeta]) GetHost() string { return f.host }

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
func (f *Forge[tMeta]) AddTrip(chainTrips ...ChainTrip) *Forge[tMeta] {
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
