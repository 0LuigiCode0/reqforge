package reqforge

import (
	stdcontext "context"
	stdio "io"
	"net/http"
	stdstrings "strings"

	"github.com/0LuigiCode0/errors"
)

// ChainTrip — middleware: обёртка над следующим Tripper-ом.
type ChainTrip func(next Trip) Trip

// Trip — один шаг цепочки: запрос/ответ или вызов следующего Tripper-а.
type Trip func(req *http.Request) (*http.Response, error)

// wrapBaseTrip — конец цепочки: вызывает Meta.BaseTrip из контекста.
func wrapBaseTrip[tMeta any](req *http.Request) (*http.Response, error) {
	m := MetaFromCtx[tMeta](req.Context())
	if m.BaseTrip == nil {
		return nil, errors.New("meta baseTrip not found")
	}

	return m.BaseTrip(req)
}

// tripper — адаптер http.RoundTripper для цепочки Trip.
type tripper[tMeta any] struct {
	trip Trip
	// meta — заполняется в Request.Do, но не используется: Meta достаётся из контекста.
	meta *Meta[tMeta]
}

func (t *tripper[tMeta]) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.trip(req)
}

// TimeoutTrip — таймаут из meta.Timeout (если > 0) на контекст запроса.
func (*Forge[tMeta]) TimeoutTrip(next Trip) Trip {
	return func(req *http.Request) (*http.Response, error) {
		meta := MetaFromCtx[tMeta](req.Context())

		if meta.Timeout > 0 {
			ctxtime, cancel := stdcontext.WithTimeout(req.Context(), meta.Timeout)
			defer cancel()
			return next(req.WithContext(ctxtime))
		}

		return next(req)
	}
}

// NginxTrip — не-OK ответ с текстовым Content-Type (страница ошибки NGINX) → ошибка с телом:
// downstream-ретрай сможет не ретраить мёртвый прокси (502/504).
func (*Forge[tMeta]) NginxTrip(next Trip) Trip {
	return func(req *http.Request) (*http.Response, error) {
		resp, err := next(req)
		if err != nil {
			return nil, err
		}

		content := resp.Header.Get("Content-Type")
		if resp.StatusCode != http.StatusOK && (stdstrings.HasPrefix(content, "text/html") || stdstrings.HasPrefix(content, "text/plain")) {
			dump, err := stdio.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			_ = resp.Body.Close()

			return nil, errors.New("NGINX " + string(dump))
		}

		return resp, nil
	}
}
