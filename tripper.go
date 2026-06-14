package reqforge

import "net/http"

type ChainTrip[t_meta any] func(next Trip[t_meta]) Trip[t_meta]

type Trip[t_meta any] func(ctx *Ctx[t_meta], req *http.Request) (*http.Response, error)

func wrapBaseTrip[t_meta any](ctx *Ctx[t_meta], req *http.Request) (*http.Response, error) {
	return ctx.tripper.RoundTrip(req)
}
