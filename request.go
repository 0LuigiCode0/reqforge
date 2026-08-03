package reqforge

import (
	"context"
	"net/http"
	"net/url"
	"slices"
)

type Request[tIn any, tOut, tMeta any] struct {
	core[tMeta]

	url    string
	method Method

	encoder encoder[tIn, tMeta]
	decoder decoder[*tOut, tMeta]
}

func NewRequest[tIn any, tOut, tMeta any](f *Forge[tMeta], method Method, path string) *Request[tIn, tOut, tMeta] {
	return &Request[tIn, tOut, tMeta]{
		url:    f.host + path,
		method: method,
		core:   f.core.clone(),
	}
}

func (r *Request[tIn, tOut, tMeta]) SetHeader(header http.Header) *Request[tIn, tOut, tMeta] {
	r.setHeader(header)
	return r
}

func (r *Request[tIn, tOut, tMeta]) SetContentType(contentType ContentType) *Request[tIn, tOut, tMeta] {
	r.setContentType(contentType)
	return r
}

func (r *Request[tIn, tOut, tMeta]) SetUserAgent(userAgent string) *Request[tIn, tOut, tMeta] {
	r.setUserAgent(userAgent)
	return r
}

func (r *Request[tIn, tOut, tMeta]) SetUrlValue(urlValues url.Values) *Request[tIn, tOut, tMeta] {
	r.setUrlValue(urlValues)
	return r
}

func (r *Request[tIn, tOut, tMeta]) AddTrip(chainTrips ...ChainTrip[tMeta]) *Request[tIn, tOut, tMeta] {
	r.addTrip(chainTrips)
	return r
}

func (r *Request[tIn, tOut, tMeta]) AddHeader(header http.Header) *Request[tIn, tOut, tMeta] {
	r.addHeader(header)
	return r
}

func (r *Request[tIn, tOut, tMeta]) AddUrlValue(urlValues url.Values) *Request[tIn, tOut, tMeta] {
	r.addUrlValue(urlValues)
	return r
}

func (r *Request[tIn, tOut, tMeta]) SetEncoder(encoder encoder[tIn, tMeta]) *Request[tIn, tOut, tMeta] {
	r.encoder = encoder
	return r
}

func (r *Request[tIn, tOut, tMeta]) SetDecoder(decoder decoder[*tOut, tMeta]) *Request[tIn, tOut, tMeta] {
	r.decoder = decoder
	return r
}

func (r *Request[tIn, tOut, tMeta]) Do(ctx context.Context, in tIn, meta *tMeta) (*tOut, error) {
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

	reqCtx.Header = make(http.Header, len(r.header))
	for k, v := range r.header {
		reqCtx.Header[k] = slices.Clone(v)
	}
	urlValues := make(url.Values, len(r.urlValues))
	for k, v := range r.urlValues {
		urlValues[k] = slices.Clone(v)
	}
	reqCtx.URL.RawQuery = urlValues.Encode()

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
