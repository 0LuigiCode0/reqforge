package reqforge

import (
	"maps"
	"net/http"
	"net/url"
)

type core[t_meta any] struct {
	header    http.Header
	urlValues url.Values

	baseTrip http.RoundTripper
	trip     Trip[t_meta]
}

func (c *core[t_meta]) setHeader(header http.Header) {
	c.header = header
}

func (c *core[t_meta]) setContentType(contentType ContentType) {
	c.header.Set(HeaderContentType, contentType)
}

func (c *core[t_meta]) setUserAgent(userAgent string) {
	c.header.Set(HeaderUserAgent, userAgent)
}

func (c *core[t_meta]) setUrlValue(urlValues url.Values) {
	c.urlValues = urlValues
}

func (c *core[t_meta]) addTrip(chainTrips []ChainTrip[t_meta]) {
	for _, chainTrip := range chainTrips {
		c.trip = chainTrip(c.trip)
	}
}

func (c *core[t_meta]) addHeader(header http.Header) {
	for k, v := range header {
		c.header[k] = append(c.header[k], v...)
	}
}

func (c *core[t_meta]) addUrlValue(urlValues url.Values) {
	for k, v := range urlValues {
		c.urlValues[k] = append(c.urlValues[k], v...)
	}
}

func (c *core[t_meta]) clone() core[t_meta] {
	return core[t_meta]{
		header:    maps.Clone(c.header),
		urlValues: maps.Clone(c.urlValues),
		baseTrip:  c.baseTrip,
		trip:      c.trip,
	}
}
