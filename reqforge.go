package reqforge

import (
	"net/http"
	"net/url"
)

type Forge[tMeta any] struct {
	core[tMeta]

	host string
}

func NewForge[tMeta any](host string, baseTrip http.RoundTripper) *Forge[tMeta] {
	return &Forge[tMeta]{
		host: host,
		core: core[tMeta]{
			baseTrip:  baseTrip,
			trip:      wrapBaseTrip[tMeta],
			header:    http.Header{},
			urlValues: url.Values{},
		},
	}
}

func (f *Forge[tMeta]) SetHeader(header http.Header) *Forge[tMeta] {
	f.setHeader(header)
	return f
}

func (f *Forge[tMeta]) SetContentType(contentType ContentType) *Forge[tMeta] {
	f.setContentType(contentType)
	return f
}

func (f *Forge[tMeta]) SetUserAgent(userAgent string) *Forge[tMeta] {
	f.setUserAgent(userAgent)
	return f
}

func (f *Forge[tMeta]) SetUrlValue(urlValues url.Values) *Forge[tMeta] {
	f.setUrlValue(urlValues)
	return f
}

func (f *Forge[tMeta]) AddTrip(chainTrips ...ChainTrip[tMeta]) *Forge[tMeta] {
	f.addTrip(chainTrips)
	return f
}

func (f *Forge[tMeta]) AddHeader(header http.Header) *Forge[tMeta] {
	f.addHeader(header)
	return f
}

func (f *Forge[tMeta]) AddUrlValue(urlValues url.Values) *Forge[tMeta] {
	f.addUrlValue(urlValues)
	return f
}

func (f *Forge[tMeta]) SetEncoder(encoder encoder[any, tMeta]) *Forge[tMeta] {
	f.setEncoder(encoder)
	return f
}

func (f *Forge[tMeta]) SetDecoder(decoder decoder[any, tMeta]) *Forge[tMeta] {
	f.setDecoder(decoder)
	return f
}
