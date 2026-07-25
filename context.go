package reqforge

import "net/http"

type Ctx[t_meta comparable] struct {
	tripper http.RoundTripper

	Meta *t_meta
}

func (ctx *Ctx[t_meta]) SetBaseTrip(tripper http.RoundTripper) {
	ctx.tripper = tripper
}
