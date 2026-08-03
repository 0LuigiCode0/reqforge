package reqforge

import "net/http"

// Ctx — контекст middleware-цепочки, передаётся в каждую функцию Trip.
// Содержит доступ к base http.RoundTripper и произвольные метаданные.
type Ctx[tMeta any] struct {
	tripper http.RoundTripper

	Meta *tMeta
}

// SetBaseTrip — изменить базовый RoundTripper в контексте.
// Используется middleware для подмены транспорта (например, для моков).
func (ctx *Ctx[tMeta]) SetBaseTrip(tripper http.RoundTripper) {
	ctx.tripper = tripper
}
