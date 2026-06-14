package reqforge

import "net/http"

type Ctx[t_meta any] struct {
	tripper http.RoundTripper

	Meta *t_meta
}

func (ctx *Ctx[t_meta]) SetBaseTrip(tripper http.RoundTripper) {
	ctx.tripper = tripper
}
