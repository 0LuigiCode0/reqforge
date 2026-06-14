package reqforge

import (
	"net/http"
	"net/url"
)

type Request[t_req, t_resp, t_meta any] struct {
	url    string
	method Method

	core[t_meta]
}

func NewRequest[t_req, t_resp, t_meta any](f *Forge[t_meta], path string, method Method) *Request[t_req, t_resp, t_meta] {
	return &Request[t_req, t_resp, t_meta]{
		url:    f.host + path,
		method: method,
		core:   f.core.clone(),
	}
}

func (r *Request[t_req, t_resp, t_meta]) SetHeader(header http.Header) *Request[t_req, t_resp, t_meta] {
	r.setHeader(header)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) SetContentType(contentType ContentType) *Request[t_req, t_resp, t_meta] {
	r.setContentType(contentType)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) SetUserAgent(userAgent string) *Request[t_req, t_resp, t_meta] {
	r.setUserAgent(userAgent)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) SetUrlValue(urlValues url.Values) *Request[t_req, t_resp, t_meta] {
	r.setUrlValue(urlValues)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) AddTrip(chainTrips ...ChainTrip[t_meta]) *Request[t_req, t_resp, t_meta] {
	r.addTrip(chainTrips)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) AddHeader(header http.Header) *Request[t_req, t_resp, t_meta] {
	r.addHeader(header)
	return r
}

func (r *Request[t_req, t_resp, t_meta]) AddUrlValue(urlValues url.Values) *Request[t_req, t_resp, t_meta] {
	r.addUrlValue(urlValues)
	return r
}
