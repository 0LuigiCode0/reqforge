package reqforge

import (
	stdcontext "context"
	stdtime "time"
)

// Meta — метаданные запроса, передаваемые в middleware-цепочку.
type Meta[tMeta any] struct {
	// BaseTrip — если nil, в Request.Do подставляется baseTrip из Forge.
	BaseTrip Trip
	// Procedure — имя процедуры/endpoint для трассировки.
	Procedure string

	// LogEnable — если true, логирование ответа форсируется (по умолчанию — зависит от DebugEnable).
	LogEnable bool
	// FullLog — если true, большие ответы не обрезаются при логировании (TrimSize пропускает обрезку).
	FullLog bool
	// Timeout — таймаут выполнения запроса (используется TimeoutTrip).
	Timeout stdtime.Duration
	// Attempt — номер текущей попытки (инкрементируется middleware ретраев).
	Attempt int

	// Ext — произвольные дополнительные метаданные для middleware.
	Ext *tMeta
}

// NewMeta — метаданные для middleware-цепочки с применением опций.
func NewMeta[tMeta any](procedure string, opt ...Option[tMeta]) *Meta[tMeta] {
	m := &Meta[tMeta]{
		Procedure: procedure,
		Ext:       new(tMeta),
	}
	for _, v := range opt {
		v(m)
	}

	return m
}

type metaKey struct{}

func (s *Meta[tMeta]) WithCtx(ctx stdcontext.Context) stdcontext.Context {
	return stdcontext.WithValue(ctx, metaKey{}, s)
}

// MetaFromCtx — извлекает Meta из контекста (если её нет, создаёт пустую).
func MetaFromCtx[tMeta any](ctx stdcontext.Context) *Meta[tMeta] {
	v, ok := ctx.Value(metaKey{}).(*Meta[tMeta])
	if !ok {
		v = NewMeta[tMeta]("undefined")
	}

	return v
}

// Option — опция конфигурации метаданных.
type Option[tMeta any] func(meta *Meta[tMeta])

// WithCustomTrip — подмена транспорта (например, для моков).
func WithCustomTrip[tMeta any](trip Trip) Option[tMeta] {
	return func(meta *Meta[tMeta]) { meta.BaseTrip = trip }
}

// WithTimeout — таймаут для запроса.
func WithTimeout[tMeta any](timeout stdtime.Duration) Option[tMeta] {
	return func(m *Meta[tMeta]) { m.Timeout = timeout }
}

// WithLog — форс логирование ответа для этого запроса.
func WithLog[tMeta any]() Option[tMeta] {
	return func(m *Meta[tMeta]) { m.LogEnable = true }
}

// WithFullLog — отключает обрезку больших ответов при логировании (TrimSize).
func WithFullLog[tMeta any]() Option[tMeta] {
	return func(m *Meta[tMeta]) { m.FullLog = true }
}
