package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// -------------------------------------------------------------------------- //
//  Shared types — общие для клиента и сервера
// -------------------------------------------------------------------------- //

// Hello — уникальный идентификатор запроса (key для sync.Map)
type Hello int64

// RequestBody — тело запроса, отправляемое клиентом
type RequestBody struct {
	ID      Hello  `json:"id"`
	Payload string `json:"payload"`
}

// ResponseBody — тело ответа от сервера
type ResponseBody struct {
	Status string `json:"status"`
}

// -------------------------------------------------------------------------- //
//  Global state
// -------------------------------------------------------------------------- //

var (
	// processed — map обработанных запросов: key=ID, value=RequestBody
	processed sync.Map
	wg        sync.WaitGroup
	// retryCounter — счётчик для retry endpoint
	retryCounter int64
)

// -------------------------------------------------------------------------- //
//  sendRequests — универсальная отправка запросов
// -------------------------------------------------------------------------- //

type requestFunc func(ctx context.Context, id Hello, payload string) (*ResponseBody, error)

func sendRequests(ctx context.Context, label string, offset int, prefix string, num int, fn requestFunc) {
	fmt.Printf("[main] отправляем %d запросов (%s) …\n", num, label)
	fmt.Println()

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			id := Hello(offset + idx + 1)
			payload := fmt.Sprintf("%s-%d", prefix, idx+1)

			result, err := fn(ctx, id, payload)
			if err != nil {
				fmt.Printf("[client-%d] failed: %v\n", idx, err)
				return
			}

			fmt.Printf("[client-%d] status=%q — удаляем ключ %v из map\n", idx, result.Status, id)
			if _, loaded := processed.Load(id); loaded {
				processed.Delete(id)
			} else {
				fmt.Printf("[client-%d] ключ %v уже удалён\n", idx, id)
			}
		}(i)
	}
}

// sendSingleRequest — отправка одного запроса с обработкой результата
func sendSingleRequest(ctx context.Context, label string, id Hello, payload string, fn requestFunc) {
	fmt.Printf("[main] %s …\n", label)
	fmt.Println()

	wg.Add(1)
	go func() {
		defer wg.Done()

		result, err := fn(ctx, id, payload)
		if err != nil {
			fmt.Printf("[client] failed: %v\n", err)
			return
		}

		fmt.Printf("[client] status=%q — удаляем ключ %v из map\n", result.Status, id)
		if _, loaded := processed.Load(id); loaded {
			processed.Delete(id)
		} else {
			fmt.Printf("[client] ключ %v уже удалён\n", id)
		}
	}()
}

// -------------------------------------------------------------------------- //
//  main — оркестрация
// -------------------------------------------------------------------------- //

func main() {
	fmt.Println("=== reqforge client-server example ===")
	fmt.Println()

	const numRequests = 10

	// --- 1. Запуск сервера ---
	fmt.Println("[main] запускаем сервер на :8080 …")

	handlers := map[string]http.HandlerFunc{
		"/process":           processEndpoint,
		"/process-form":      processFormEndpoint,
		"/process-multipart": processMultipartEndpoint,
		"/process-retry":     processRetryEndpoint,
		"/process-mock":      processMockEndpoint,
		"/echo":              echoEndpoint,
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           setupServer(handlers),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("[main] сервер упал: %v\n", err)
		}
	}()

	// --- 2. Создаём клиент (Forge + эндпоинты скрыты внутри) ---
	fmt.Println("[main] создаём клиент …")
	client := NewClient()

	// --- 3. Health check ---
	fmt.Println("[main] ожидаем готовности сервера …")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.HealthCheck(ctx); err != nil {
		fmt.Printf("[main] сервер не ответил: %v — выход\n", err)
		server.Close()
		return
	}
	fmt.Println("[main] сервер готов!")
	fmt.Println()

	// --- 4. Отправка запросов ---
	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Batch-запросы
	sendRequests(ctx, "JSON", 0, "payload", numRequests, client.Process)
	sendRequests(ctx, "query params", numRequests, "form-payload", numRequests, client.ProcessForm)
	sendRequests(ctx, "multipart/form-data", numRequests*2, "multipart-payload", numRequests, client.ProcessMultipart)

	// Single-запросы (retry + mock)
	sendSingleRequest(ctx, "1 retry-запрос (middleware: AddTrip)", Hello(numRequests*3+1), "retry-payload", client.ProcessRetry)
	sendSingleRequest(ctx, "1 mock-запрос (SetBaseTrip → mockTransport)", Hello(numRequests*3+2), "mock-payload", client.ProcessMock)

	fmt.Println("[main] ожидаем завершения всех goroutine …")
	wg.Wait()
	fmt.Println("[main] все goroutine завершены!")

	// --- 5. Демонстрация копирования и изоляции (Forge → Request) ---

	resp1, err := client.Header(ctx)
	if err != nil {
		fmt.Printf("[main] ошибка DemoReq1: %v\n", err)
	}
	resp2, err := client.Query(ctx)
	if err != nil {
		fmt.Printf("[main] ошибка DemoReq2: %v\n", err)
	}

	fmt.Println()
	fmt.Println("=== Проверка header/query ===")

	if reflect.DeepEqual(resp1, &EchoResponse{
		Headers: map[string]string{
			"Accept-Encoding":   "gzip",
			"User-Agent":        "Go-http-client/1.1",
			"X-Req1":            "yes",
			"X-Req2":            "hello",
			"X-Template-Header": "shared",
		},
		Query: map[string]string{
			"template_param": "from_template",
		},
	}) {
		fmt.Println("header valid")
	} else {
		fmt.Println("header invalid")
	}
	if reflect.DeepEqual(resp2, &EchoResponse{
		Headers: map[string]string{
			"Accept-Encoding":   "gzip",
			"User-Agent":        "Go-http-client/1.1",
			"X-Template-Header": "shared",
		},
		Query: map[string]string{
			"req2_param_1":   "unique",
			"req2_param_2":   "world",
			"template_param": "from_template",
		},
	}) {
		fmt.Println("query valid")
	} else {
		fmt.Println("query invalid")
	}

	// --- 6. Проверка ---

	fmt.Println()
	fmt.Println("=== Проверка sync.Map ===")

	count := 0
	processed.Range(func(key, value any) bool {
		count++
		fmt.Printf("  осталось: key=%v\n", key)
		return true
	})

	totalRequests := numRequests*3 + 2
	if count == 0 {
		fmt.Println()
		fmt.Printf("✓ map пуст — все %d запросов успешно обработаны и удалены\n", totalRequests)
		fmt.Println("✓ ТЕСТ PASSED")
	} else {
		fmt.Println()
		fmt.Printf("✗ map НЕ пуст — осталось %d из %d запросов\n", count, totalRequests)
		fmt.Println("✗ ТЕСТ FAILED")
	}

	// Сброс счётчика retry для чистоты
	atomic.StoreInt64(&retryCounter, 0)

	server.Close()
}
