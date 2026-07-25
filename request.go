package reqforge

import (
	"context"
	"net/http"
	"net/url"
)

type Request[t_in, t_out, t_meta comparable] struct {
	core[t_meta]

	url    string
	method Method

	encoder encoder[*t_in, t_meta]
	decoder decoder[*t_out, t_meta]
}

func NewRequest[t_in, t_out, t_meta comparable](f *Forge[t_meta], path string, method Method) *Request[t_in, t_out, t_meta] {
	return &Request[t_in, t_out, t_meta]{
		url:    f.host + path,
		method: method,
		core:   f.core.clone(),
	}
}

func (r *Request[t_in, t_out, t_meta]) SetHeader(header http.Header) *Request[t_in, t_out, t_meta] {
	r.setHeader(header)
	return r
}

func (r *Request[t_in, t_out, t_meta]) SetContentType(contentType ContentType) *Request[t_in, t_out, t_meta] {
	r.setContentType(contentType)
	return r
}

func (r *Request[t_in, t_out, t_meta]) SetUserAgent(userAgent string) *Request[t_in, t_out, t_meta] {
	r.setUserAgent(userAgent)
	return r
}

func (r *Request[t_in, t_out, t_meta]) SetUrlValue(urlValues url.Values) *Request[t_in, t_out, t_meta] {
	r.setUrlValue(urlValues)
	return r
}

func (r *Request[t_in, t_out, t_meta]) AddTrip(chainTrips ...ChainTrip[t_meta]) *Request[t_in, t_out, t_meta] {
	r.addTrip(chainTrips)
	return r
}

func (r *Request[t_in, t_out, t_meta]) AddHeader(header http.Header) *Request[t_in, t_out, t_meta] {
	r.addHeader(header)
	return r
}

func (r *Request[t_in, t_out, t_meta]) AddUrlValue(urlValues url.Values) *Request[t_in, t_out, t_meta] {
	r.addUrlValue(urlValues)
	return r
}

func (r *Request[t_in, t_out, t_meta]) SetEncoder(encoder encoder[*t_in, t_meta]) *Request[t_in, t_out, t_meta] {
	r.encoder = encoder
	return r
}

func (r *Request[t_in, t_out, t_meta]) SetDecoder(decoder decoder[*t_out, t_meta]) *Request[t_in, t_out, t_meta] {
	r.decoder = decoder
	return r
}

func (r *Request[t_in, t_out, t_meta]) Do(ctx context.Context, in *t_in) (*t_out, error) {
	tripper := &tripper[t_meta]{
		ctx: &Ctx[t_meta]{
			Meta:    new(t_meta),
			tripper: r.baseTrip,
		},
		trip: r.trip,
	}

	reqCtx, err := http.NewRequestWithContext(ctx, r.method, r.url, nil)
	if err != nil {
		return nil, err
	}

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

	out := new(t_out)
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
