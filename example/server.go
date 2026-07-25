package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// -------------------------------------------------------------------------- //
//  ServerMiddleware — аналог ChainTrip для серверного HTTP-обработчика
// -------------------------------------------------------------------------- //

type ServerMiddleware func(next http.Handler) http.Handler

// ChainMiddleware строит цепочку: последний middleware — outermost
func ChainMiddleware(middlewares []ServerMiddleware, final http.Handler) http.Handler {
	h := final
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// -------------------------------------------------------------------------- //
//  setupServer — строит middleware-цепочку и возвращает http.Handler
// -------------------------------------------------------------------------- //

func setupServer(handlers map[string]http.HandlerFunc) http.Handler {
	// logging middleware — логирование метода, пути, статуса, времени
	logging := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			fmt.Printf("[server] %s %s %d %s\n", r.Method, r.URL.Path, sw.status, time.Since(start))
		})
	}

	// recovery middleware — panic recovery → 500
	recovery := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					fmt.Printf("[server] PANIC: %v\n", rec)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}

	chain := []ServerMiddleware{logging, recovery}

	// default handler — 404
	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	})

	// route handler — маршрутизация по URL.Path
	routeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := handlers[r.URL.Path]; ok {
			handler(w, r)
		} else {
			defaultHandler.ServeHTTP(w, r)
		}
	})

	return ChainMiddleware(chain, routeHandler)
}

// -------------------------------------------------------------------------- //
//  statusWriter — обёртка для перехвата HTTP status code
// -------------------------------------------------------------------------- //

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// -------------------------------------------------------------------------- //
//  processEndpoint — обработка POST /process
// -------------------------------------------------------------------------- //

func processEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		fmt.Printf("[server] 400 Bad Request: %v\n", err)
		return
	}

	processed.Store(body.ID, body)
	fmt.Printf("[server] обработан id=%v payload=%q\n", body.ID, body.Payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Status: "processed"})
}

// -------------------------------------------------------------------------- //
//  processFormEndpoint — обработка POST /process-form (query params)
// -------------------------------------------------------------------------- //

func processFormEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query params — body = nil
	idStr := r.URL.Query().Get("id")
	payload := r.URL.Query().Get("payload")

	var body RequestBody
	if _, err := fmt.Sscanf(idStr, "%d", &body.ID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		fmt.Printf("[server] 400 Bad Request: invalid id=%q\n", idStr)
		return
	}
	body.Payload = payload

	processed.Store(body.ID, body)
	fmt.Printf("[server] обработан query id=%v payload=%q\n", body.ID, body.Payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Status: "processed"})
}

// -------------------------------------------------------------------------- //
//  processMultipartEndpoint — обработка POST /process-multipart (multipart/form-data)
// -------------------------------------------------------------------------- //

func processMultipartEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		fmt.Printf("[server] 400 Bad Request: %v\n", err)
		return
	}

	idStr := r.FormValue("id")
	payload := r.FormValue("payload")

	var body RequestBody
	if _, err := fmt.Sscanf(idStr, "%d", &body.ID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		fmt.Printf("[server] 400 Bad Request: invalid id=%q\n", idStr)
		return
	}
	body.Payload = payload

	processed.Store(body.ID, body)
	fmt.Printf("[server] обработан multipart id=%v payload=%q\n", body.ID, body.Payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Status: "processed"})
}

// -------------------------------------------------------------------------- //
//  processRetryEndpoint — эндпоинт для демонстрации retry (возвращает 500 → retry)
// -------------------------------------------------------------------------- //

func processRetryEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		return
	}

	// Первая попытка — 500, вторая — 200
	count := atomic.AddInt64(&retryCounter, 1)
	if count == 1 {
		fmt.Printf("[server] retry endpoint: attempt %d → 500\n", count)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "temporary failure"})
		return
	}

	fmt.Printf("[server] retry endpoint: attempt %d → 200\n", count)
	processed.Store(body.ID, body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Status: "retry-processed"})
}

// -------------------------------------------------------------------------- //
//  processMockEndpoint — эндпоинт для демонстрации mock (всегда 200)
// -------------------------------------------------------------------------- //

func processMockEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Status: "mock-processed"})
}
