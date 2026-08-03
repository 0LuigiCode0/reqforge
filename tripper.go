package reqforge

import "net/http"

type ChainTrip[tMeta any] func(next Trip[tMeta]) Trip[tMeta]

type Trip[tMeta any] func(ctx *Ctx[tMeta], req *http.Request) (*http.Response, error)

func wrapBaseTrip[tMeta any](ctx *Ctx[tMeta], req *http.Request) (*http.Response, error) {
	return ctx.tripper.RoundTrip(req)
}

type tripper[tMeta any] struct {
	trip Trip[tMeta]
	ctx  *Ctx[tMeta]
}

func (t *tripper[tMeta]) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.trip(t.ctx, r)
}
