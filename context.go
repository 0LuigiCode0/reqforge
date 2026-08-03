package reqforge

import "net/http"

type Ctx[tMeta any] struct {
	tripper http.RoundTripper

	Meta *tMeta
}

func (ctx *Ctx[tMeta]) SetBaseTrip(tripper http.RoundTripper) {
	ctx.tripper = tripper
}
