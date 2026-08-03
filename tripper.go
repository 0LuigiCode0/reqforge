package reqforge

import "net/http"

// ChainTrip — тип middleware-цепочки. Принимает следующую Trip и возвращает новую Trip.
// Позволяет оборачивать запрос/ответ (логирование, retry, timeout и т.д.).
type ChainTrip[tMeta any] func(next Trip[tMeta]) Trip[tMeta]

// Trip — функция, выполняющая один шаг в цепочке middleware.
// Вызывает следующую Trip или базовый RoundTripper.
type Trip[tMeta any] func(ctx *Ctx[tMeta], req *http.Request) (*http.Response, error)

// wrapBaseTrip — конечный элемент цепочки, вызывает базовый http.RoundTripper из контекста.
func wrapBaseTrip[tMeta any](ctx *Ctx[tMeta], req *http.Request) (*http.Response, error) {
	return ctx.tripper.RoundTrip(req)
}

// tripper — адаптер http.RoundTripper, который выполняет цепочку Trip через контекст.
type tripper[tMeta any] struct {
	trip Trip[tMeta]
	ctx  *Ctx[tMeta]
}

func (t *tripper[tMeta]) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.trip(t.ctx, r)
}
