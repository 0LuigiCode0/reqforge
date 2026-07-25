package reqforge

import (
	"net/http"
	"net/url"
)

type Forge[t_meta comparable] struct {
	core[t_meta]

	host string
}

func NewForge[t_meta comparable](host string, baseTrip http.RoundTripper) *Forge[t_meta] {
	return &Forge[t_meta]{
		host: host,
		core: core[t_meta]{
			baseTrip:  baseTrip,
			trip:      wrapBaseTrip[t_meta],
			header:    http.Header{},
			urlValues: url.Values{},
		},
	}
}

func (f *Forge[t_meta]) SetHeader(header http.Header) *Forge[t_meta] {
	f.setHeader(header)
	return f
}

func (f *Forge[t_meta]) SetContentType(contentType ContentType) *Forge[t_meta] {
	f.setContentType(contentType)
	return f
}

func (f *Forge[t_meta]) SetUserAgent(userAgent string) *Forge[t_meta] {
	f.setUserAgent(userAgent)
	return f
}

func (f *Forge[t_meta]) SetUrlValue(urlValues url.Values) *Forge[t_meta] {
	f.setUrlValue(urlValues)
	return f
}

func (f *Forge[t_meta]) AddTrip(chainTrips ...ChainTrip[t_meta]) *Forge[t_meta] {
	f.addTrip(chainTrips)
	return f
}

func (f *Forge[t_meta]) AddHeader(header http.Header) *Forge[t_meta] {
	f.addHeader(header)
	return f
}

func (f *Forge[t_meta]) AddUrlValue(urlValues url.Values) *Forge[t_meta] {
	f.addUrlValue(urlValues)
	return f
}

func (f *Forge[t_meta]) SetEncoder(encoder encoder[any, t_meta]) *Forge[t_meta] {
	f.setEncoder(encoder)
	return f
}

func (f *Forge[t_meta]) SetDecoder(decoder decoder[any, t_meta]) *Forge[t_meta] {
	f.setDecoder(decoder)
	return f
}
