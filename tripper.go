package reqforge

import "net/http"

type ChainTrip[t_meta comparable] func(next Trip[t_meta]) Trip[t_meta]

type Trip[t_meta comparable] func(ctx *Ctx[t_meta], req *http.Request) (*http.Response, error)

func wrapBaseTrip[t_meta comparable](ctx *Ctx[t_meta], req *http.Request) (*http.Response, error) {
	return ctx.tripper.RoundTrip(req)
}

type tripper[t_meta comparable] struct {
	trip Trip[t_meta]
	ctx  *Ctx[t_meta]
}

func (t *tripper[t_meta]) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.trip(t.ctx, r)
}
