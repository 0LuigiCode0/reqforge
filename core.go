package reqforge

import (
	"net/http"
	"net/url"
	"slices"
)

type core[t_meta comparable] struct {
	header    http.Header
	urlValues url.Values

	baseTrip http.RoundTripper
	trip     Trip[t_meta]

	encoder encoder[any, t_meta]
	decoder decoder[any, t_meta]
}

type (
	encoder[t_value, t_meta comparable] func(ctx *Ctx[t_meta], v t_value, req *http.Request) error
	decoder[t_value, t_meta comparable] func(ctx *Ctx[t_meta], v t_value, resp *http.Response) error
)

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

func (c *core[t_meta]) setEncoder(encoder encoder[any, t_meta]) {
	c.encoder = encoder
}

func (c *core[t_meta]) setDecoder(decoder decoder[any, t_meta]) {
	c.decoder = decoder
}

// -------------------------------------------------------------------------- //

func (c *core[t_meta]) clone() core[t_meta] {
	header := make(http.Header, len(c.header))
	for k, v := range c.header {
		header[k] = slices.Clone(v)
	}

	urlValues := make(url.Values, len(c.urlValues))
	for k, v := range c.urlValues {
		urlValues[k] = slices.Clone(v)
	}

	return core[t_meta]{
		header:    header,
		urlValues: urlValues,
		baseTrip:  c.baseTrip,
		trip:      c.trip,
		encoder:   c.encoder,
		decoder:   c.decoder,
	}
}
