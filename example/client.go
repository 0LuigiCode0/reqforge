package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	reqforge "github.com/0LuigiCode0/reqforge"
)

// -------------------------------------------------------------------------- //
//  AppMeta — пользовательский контекст, передаваемый через middleware
// -------------------------------------------------------------------------- //

type AppMeta struct {
	RequestID int64
	Timestamp time.Time
}

// -------------------------------------------------------------------------- //
//  Client — публичный интерфейс клиентской части
//  Все внутренние детали (Forge, Request, encoder/decoder) скрыты
// -------------------------------------------------------------------------- //

type Client struct {
	forge            *reqforge.Forge[AppMeta]
	process          *reqforge.Request[RequestBody, ResponseBody, AppMeta]
	processForm      *reqforge.Request[RequestBody, ResponseBody, AppMeta]
	processMultipart *reqforge.Request[RequestBody, ResponseBody, AppMeta]
	processRetry     *reqforge.Request[RequestBody, ResponseBody, AppMeta]
	processMock      *reqforge.Request[RequestBody, ResponseBody, AppMeta]
	health           *reqforge.Request[struct{}, HealthResponse, AppMeta]
}

// HealthResponse — тело ответа health check
type HealthResponse struct {
	Status string `json:"status"`
}

// -------------------------------------------------------------------------- //
//  Middleware — именованные функции, а не лямбды
// -------------------------------------------------------------------------- //

// loggingTrip — логирование метода, пути, статуса, времени
func loggingTrip[T_meta comparable](next reqforge.Trip[T_meta]) reqforge.Trip[T_meta] {
	return func(c *reqforge.Ctx[T_meta], req *http.Request) (*http.Response, error) {
		fmt.Printf("[client] %s %s\n", req.Method, req.URL.Path)
		start := time.Now()
		resp, err := next(c, req)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[client] error: %v (%v)\n", err, elapsed)
		} else {
			fmt.Printf("[client] %d (%v)\n", resp.StatusCode, elapsed)
		}
		return resp, err
	}
}

// timeoutTrip — устанавливает таймаут 5s через context.WithTimeout
func timeoutTrip[T_meta comparable](next reqforge.Trip[T_meta]) reqforge.Trip[T_meta] {
	return func(c *reqforge.Ctx[T_meta], req *http.Request) (*http.Response, error) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return next(c, req.Clone(timeoutCtx))
	}
}

// -------------------------------------------------------------------------- //
//  retryTrip — middleware retry (AddTrip, НЕ SetBaseTrip)
// -------------------------------------------------------------------------- //

// retryTrip — retry через middleware: пробует 3 раза при 5xx
func retryTrip[T_meta comparable](next reqforge.Trip[T_meta]) reqforge.Trip[T_meta] {
	return func(c *reqforge.Ctx[T_meta], req *http.Request) (*http.Response, error) {
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			// Сбрасываем body перед каждой попыткой
			if req.Body != nil && req.GetBody != nil {
				body, err := req.GetBody()
				if err == nil {
					req.Body = body
				}
			}
			resp, err := next(c, req)
			if err != nil {
				lastErr = err
				continue // retry on error
			}
			// Если 5xx — retry
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				lastErr = fmt.Errorf("server error: %d (attempt %d)", resp.StatusCode, attempt+1)
				continue
			}
			return resp, nil
		}
		return nil, lastErr
	}
}

// -------------------------------------------------------------------------- //
//  Transport replacement — пример подмены всего стека (SetBaseTrip)
//  Используется ТОЛЬКО когда нужно заменить http.RoundTripper целиком
// -------------------------------------------------------------------------- //

// mockTransport — подменяет ВЕСЬ стек: все запросы возвращают mock-ответ
// Используется для тестирования без реального сервера
type mockTransport struct {
	statusCode int
	body       string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("[client] mockTransport: %s %s → %d\n", req.Method, req.URL.Path, m.statusCode)
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(m.body))),
		Header:     make(http.Header),
	}, nil
}

// mockTransportTrip — подменяет ВЕСЬ стек на mockTransport
// ВАЖНО: SetBaseTrip используется ТОЛЬКО для подмены http.RoundTripper
func mockTransportTrip[T_meta comparable](next reqforge.Trip[T_meta]) reqforge.Trip[T_meta] {
	return func(c *reqforge.Ctx[T_meta], req *http.Request) (*http.Response, error) {
		// Подменяем ВЕСЬ транспорт на mock (возвращает 200 OK)
		c.SetBaseTrip(&mockTransport{
			statusCode: 200,
			body:       `{"status":"mock-processed"}`,
		})
		return next(c, req)
	}
}

// -------------------------------------------------------------------------- //
//  NewClient — создаёт клиент с настроенным Forge и всеми эндпоинтами
// -------------------------------------------------------------------------- //

func NewClient() *Client {
	forge := reqforge.NewForge[AppMeta](
		"http://localhost:8080",
		http.DefaultTransport,
	)

	forge = forge.AddTrip(loggingTrip) // только логирование, без timeout

	// process — 5s timeout
	process := reqforge.NewRequest[RequestBody, ResponseBody](
		forge, "/process", reqforge.MethodPost,
	).SetEncoder(jsonEncode[RequestBody]).
		SetDecoder(jsonDecode[ResponseBody]).
		AddTrip(timeoutTrip) // 5s timeout

	// processForm — 5s timeout
	processForm := reqforge.NewRequest[RequestBody, ResponseBody](
		forge, "/process-form", reqforge.MethodPost,
	).SetEncoder(queryEncodeRequestBody).
		SetDecoder(jsonDecode[ResponseBody]).
		AddTrip(timeoutTrip) // 5s timeout

	// processMultipart — 5s timeout
	processMultipart := reqforge.NewRequest[RequestBody, ResponseBody](
		forge, "/process-multipart", reqforge.MethodPost,
	).SetEncoder(multipartEncodeRequestBody).
		SetDecoder(jsonDecode[ResponseBody]).
		AddTrip(timeoutTrip) // 5s timeout

	// processRetry — retry(3x) через middleware (AddTrip, НЕ SetBaseTrip)
	processRetry := reqforge.NewRequest[RequestBody, ResponseBody](
		forge, "/process-retry", reqforge.MethodPost,
	).SetEncoder(jsonEncode[RequestBody]).
		SetDecoder(jsonDecode[ResponseBody]).
		AddTrip(retryTrip) // retry 3x при 5xx

	// processMock — подменяет ВЕСЬ стек на mockTransport (SetBaseTrip)
	// Используется для тестирования без реального сервера
	processMock := reqforge.NewRequest[RequestBody, ResponseBody](
		forge, "/process-mock", reqforge.MethodPost,
	).SetEncoder(jsonEncode[RequestBody]).
		SetDecoder(jsonDecode[ResponseBody]).
		AddTrip(mockTransportTrip) // подменяет ВЕСЬ транспорт

	health := reqforge.NewRequest[struct{}, HealthResponse](
		forge, "/", reqforge.MethodGet,
	).SetDecoder(func(c *reqforge.Ctx[AppMeta], v *HealthResponse, resp *http.Response) error {
		defer resp.Body.Close()
		v.Status = "ok"
		return nil
	})

	return &Client{
		forge:            forge,
		process:          process,
		processForm:      processForm,
		processMultipart: processMultipart,
		processRetry:     processRetry,
		processMock:      processMock,
		health:           health,
	}
}

// -------------------------------------------------------------------------- //
//  jsonEncode / jsonDecode — именованные функции-encoder/decoder
// -------------------------------------------------------------------------- //

func jsonEncode[T any](c *reqforge.Ctx[AppMeta], v *T, httpReq *http.Request) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)
	httpReq.Body = io.NopCloser(reader)
	httpReq.ContentLength = int64(len(body))
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	httpReq.Header.Set(string(reqforge.HeaderContentType), string(reqforge.ContentTypeJson))
	return nil
}

// -------------------------------------------------------------------------- //
//  queryEncodeRequestBody — сериализация RequestBody в query params (body = nil)
// -------------------------------------------------------------------------- //

func queryEncodeRequestBody(c *reqforge.Ctx[AppMeta], v *RequestBody, httpReq *http.Request) error {
	q := httpReq.URL.Query()
	q.Set("id", fmt.Sprintf("%d", v.ID))
	q.Set("payload", v.Payload)
	httpReq.URL.RawQuery = q.Encode()

	httpReq.Body = nil
	httpReq.ContentLength = 0
	return nil
}

// -------------------------------------------------------------------------- //
//  multipartEncodeRequestBody — сериализация RequestBody в multipart/form-data
// -------------------------------------------------------------------------- //

func multipartEncodeRequestBody(c *reqforge.Ctx[AppMeta], v *RequestBody, httpReq *http.Request) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormField("id")
	if err != nil {
		return err
	}
	fmt.Fprintf(fw, "%d", v.ID)

	fw, err = mw.CreateFormField("payload")
	if err != nil {
		return err
	}
	fw.Write([]byte(v.Payload))

	mw.Close()

	httpReq.Body = io.NopCloser(&buf)
	httpReq.ContentLength = int64(buf.Len())
	httpReq.Header.Set(string(reqforge.HeaderContentType), mw.FormDataContentType())
	return nil
}

func jsonDecode[T any](c *reqforge.Ctx[AppMeta], v *T, resp *http.Response) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// -------------------------------------------------------------------------- //
//  Публичные методы — только чистые названия, без деталей reqforge
// -------------------------------------------------------------------------- //

// Process отправляет запрос на обработку (JSON)
func (c *Client) Process(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	val := RequestBody{ID: id, Payload: payload}
	return c.process.Do(ctx, &val)
}

// ProcessForm отправляет запрос на обработку (form-data)
func (c *Client) ProcessForm(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	val := RequestBody{ID: id, Payload: payload}
	return c.processForm.Do(ctx, &val)
}

// ProcessMultipart отправляет запрос на обработку (multipart/form-data)
func (c *Client) ProcessMultipart(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	val := RequestBody{ID: id, Payload: payload}
	return c.processMultipart.Do(ctx, &val)
}

// ProcessRetry отправляет запрос с retry(3x) — middleware (AddTrip)
func (c *Client) ProcessRetry(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	val := RequestBody{ID: id, Payload: payload}
	return c.processRetry.Do(ctx, &val)
}

// ProcessMock отправляет запрос с mockTransport — подмена всего стека (SetBaseTrip)
func (c *Client) ProcessMock(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	val := RequestBody{ID: id, Payload: payload}
	return c.processMock.Do(ctx, &val)
}

// HealthCheck проверяет готовность сервера
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.health.Do(ctx, nil)
	return err
}
