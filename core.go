package reqforge

import (
	"net/http"
	"net/url"
	stdslices "slices"
)

// core — состояние конфигурации: заголовки, параметры URL, middleware-цепочка.
type core[tMeta any] struct {
	header    http.Header
	urlValues url.Values

	// baseTrip — fallback-RoundTripper (подставляется в meta.BaseTrip, если тот не задан).
	baseTrip Trip
	trip     Trip
}

func (c *core[tMeta]) setHeader(header http.Header) {
	c.header = header
}

func (c *core[tMeta]) setContentType(contentType ContentType) {
	c.header.Set(HeaderContentType, contentType)
}

func (c *core[tMeta]) setUserAgent(userAgent string) {
	c.header.Set(HeaderUserAgent, userAgent)
}

func (c *core[tMeta]) setUrlValue(urlValues url.Values) {
	c.urlValues = urlValues
}

func (c *core[tMeta]) addTrip(chainTrips []ChainTrip) {
	// порядок: первый ChainTrip — внешний, последний — ближайший к baseTrip
	for _, chainTrip := range stdslices.Backward(chainTrips) {
		c.trip = chainTrip(c.trip)
	}
}

func (c *core[tMeta]) addHeader(header http.Header) {
	for k, v := range header {
		c.header[k] = append(c.header[k], v...)
	}
}

func (c *core[tMeta]) addUrlValue(urlValues url.Values) {
	for k, v := range urlValues {
		c.urlValues[k] = append(c.urlValues[k], v...)
	}
}

// clone — глубокая копия для изоляции Request; trip/baseTrip — иммутабельные, копируются по ссылке.
func (c *core[tMeta]) clone() core[tMeta] {
	header := make(http.Header, len(c.header))
	for k, v := range c.header {
		header[k] = stdslices.Clone(v)
	}

	urlValues := make(url.Values, len(c.urlValues))
	for k, v := range c.urlValues {
		urlValues[k] = stdslices.Clone(v)
	}

	return core[tMeta]{
		header:    header,
		urlValues: urlValues,
		baseTrip:  c.baseTrip,
		trip:      c.trip,
	}
}
