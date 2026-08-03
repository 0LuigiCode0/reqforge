package reqforge

import (
	"net/http"
	"net/url"
	"slices"
)

type core[tMeta any] struct {
	header    http.Header
	urlValues url.Values

	baseTrip http.RoundTripper
	trip     Trip[tMeta]

	encoder encoder[any, tMeta]
	decoder decoder[any, tMeta]
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

func (c *core[tMeta]) addTrip(chainTrips []ChainTrip[tMeta]) {
	for _, chainTrip := range chainTrips {
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

func (c *core[tMeta]) setEncoder(encoder encoder[any, tMeta]) {
	c.encoder = encoder
}

func (c *core[tMeta]) setDecoder(decoder decoder[any, tMeta]) {
	c.decoder = decoder
}

// -------------------------------------------------------------------------- //

func (c *core[tMeta]) clone() core[tMeta] {
	header := make(http.Header, len(c.header))
	for k, v := range c.header {
		header[k] = slices.Clone(v)
	}

	urlValues := make(url.Values, len(c.urlValues))
	for k, v := range c.urlValues {
		urlValues[k] = slices.Clone(v)
	}

	return core[tMeta]{
		header:    header,
		urlValues: urlValues,
		baseTrip:  c.baseTrip,
		trip:      c.trip,
		encoder:   c.encoder,
		decoder:   c.decoder,
	}
}
