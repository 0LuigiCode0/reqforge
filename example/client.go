package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	process          func(context.Context, map[string]*RequestBody, *AppMeta) (*ResponseBody, error)
	processForm      func(context.Context, []RequestBody, *AppMeta) (*ResponseBody, error)
	processMultipart func(context.Context, *RequestBody, *AppMeta) (*ResponseBody, error)
	processRetry     func(context.Context, RequestBody, *AppMeta) (*ResponseBody, error)
	processMock      func(context.Context, *RequestBody, *AppMeta) (*ResponseBody, error)
	health           func(context.Context, struct{}, *AppMeta) (*HealthResponse, error)
	header           func(context.Context, string, *AppMeta) (*EchoResponse, error)
	query            func(context.Context, string, *AppMeta) (*EchoResponse, error)
}

// EchoResponse — структура ответа от /echo
type EchoResponse struct {
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
}

// HealthResponse — тело ответа health check
type HealthResponse struct {
	Status string `json:"status"`
}

// -------------------------------------------------------------------------- //
//  NewClient — создаёт клиент с настроенным Forge и всеми эндпоинтами
// -------------------------------------------------------------------------- //

func NewClient() *Client {
	c := new(Client)

	c.forge = reqforge.NewForge[AppMeta]("http://localhost:8080", http.DefaultTransport).
		SetEncoder(reqforge.JsonEncode).
		SetDecoder(reqforge.JsonDecode).
		AddTrip(loggingTrip). // только логирование, без timeout
		SetHeader(http.Header{
			"X-Template-Header": []string{"shared"},
		}).
		AddUrlValue(url.Values{
			"template_param": []string{"from_template"},
		})

	// process — 5s timeout
	c.process = reqforge.NewRequest[map[string]*RequestBody, ResponseBody](c.forge, reqforge.MethodPost, "/process").
		AddTrip(timeoutTrip).
		Do // 5s timeout

	// processForm — 5s timeout
	c.processForm = reqforge.NewRequest[[]RequestBody, ResponseBody](c.forge, reqforge.MethodPost, "/process-form").
		SetEncoder(queryEncodeRequestBody).
		AddTrip(timeoutTrip).
		Do // 5s timeout

	// processMultipart — 5s timeout
	c.processMultipart = reqforge.NewRequest[*RequestBody, ResponseBody](c.forge, reqforge.MethodPost, "/process-multipart").
		SetEncoder(multipartEncodeRequestBody).
		AddTrip(timeoutTrip).
		Do // 5s timeout

	// processRetry — retry(3x) через middleware (AddTrip, НЕ SetBaseTrip)
	c.processRetry = reqforge.NewRequest[RequestBody, ResponseBody](c.forge, reqforge.MethodPost, "/process-retry").
		AddTrip(retryTrip).
		Do // retry 3x при 5xx

	// processMock — подменяет ВЕСЬ стек на mockTransport (SetBaseTrip)
	// Используется для тестирования без реального сервера
	c.processMock = reqforge.NewRequest[*RequestBody, ResponseBody](c.forge, reqforge.MethodPost, "/process-mock").
		AddTrip(mockTransportTrip).
		Do // подменяет ВЕСЬ транспорт

	c.health = reqforge.NewRequest[struct{}, HealthResponse](c.forge, reqforge.MethodGet, "/").
		SetEncoder(reqforge.EmptyEncode).
		SetDecoder(func(c *reqforge.Ctx[AppMeta], v *HealthResponse, resp *http.Response) error {
			v.Status = "ok"
			return nil
		}).
		Do

	// demoReq1 — клон forge, добавляем уникальный заголовок
	c.header = reqforge.NewRequest[string, EchoResponse](c.forge, reqforge.MethodGet, "/echo").
		SetEncoder(func(ctx *reqforge.Ctx[AppMeta], v string, req *http.Request) error {
			req.Header.Set("X-Req2", v)
			return nil
		}).
		AddHeader(http.Header{
			"X-Req1": []string{"yes"},
		}).
		Do

	// demoReq2 — клон forge, добавляем уникальный query
	c.query = reqforge.NewRequest[string, EchoResponse](c.forge, reqforge.MethodGet, "/echo").
		SetEncoder(func(ctx *reqforge.Ctx[AppMeta], v string, req *http.Request) error {
			q := req.URL.Query()
			q.Set("req2_param_2", v)
			req.URL.RawQuery = q.Encode()
			return nil
		}).
		AddUrlValue(url.Values{
			"req2_param_1": []string{"unique"},
		}).
		Do

	return c
}

// -------------------------------------------------------------------------- //
//  Публичные методы — только чистые названия, без деталей reqforge
// -------------------------------------------------------------------------- //

// Process отправляет запрос на обработку (JSON)
func (c *Client) Process(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	return c.process(ctx, map[string]*RequestBody{"hello": {ID: id, Payload: payload}}, &AppMeta{})
}

// ProcessForm отправляет запрос на обработку (form-data)
func (c *Client) ProcessForm(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	return c.processForm(ctx, []RequestBody{{ID: id, Payload: payload}}, &AppMeta{})
}

// ProcessMultipart отправляет запрос на обработку (multipart/form-data)
func (c *Client) ProcessMultipart(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	return c.processMultipart(ctx, &RequestBody{ID: id, Payload: payload}, &AppMeta{})
}

// ProcessRetry отправляет запрос с retry(3x) — middleware (AddTrip)
func (c *Client) ProcessRetry(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	return c.processRetry(ctx, RequestBody{ID: id, Payload: payload}, &AppMeta{})
}

// ProcessMock отправляет запрос с mockTransport — подмена всего стека (SetBaseTrip)
func (c *Client) ProcessMock(ctx context.Context, id Hello, payload string) (*ResponseBody, error) {
	return c.processMock(ctx, &RequestBody{ID: id, Payload: payload}, &AppMeta{})
}

// HealthCheck проверяет готовность сервера
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.health(ctx, struct{}{}, &AppMeta{})
	return err
}

// Header выполняет запрос из клона forge с уникальным заголовком X-Req1
func (c *Client) Header(ctx context.Context) (*EchoResponse, error) {
	return c.header(ctx, "hello", &AppMeta{})
}

// Query выполняет запрос из клона forge с уникальным query param req2_param
func (c *Client) Query(ctx context.Context) (*EchoResponse, error) {
	return c.query(ctx, "world", &AppMeta{})
}

// -------------------------------------------------------------------------- //
//  Middleware — именованные функции, а не лямбды
// -------------------------------------------------------------------------- //

// loggingTrip — логирование метода, пути, статуса, времени
func loggingTrip[tMeta any](next reqforge.Trip[tMeta]) reqforge.Trip[tMeta] {
	return func(c *reqforge.Ctx[tMeta], req *http.Request) (*http.Response, error) {
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
func timeoutTrip[tMeta any](next reqforge.Trip[tMeta]) reqforge.Trip[tMeta] {
	return func(c *reqforge.Ctx[tMeta], req *http.Request) (*http.Response, error) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return next(c, req.Clone(timeoutCtx))
	}
}

// -------------------------------------------------------------------------- //
//  retryTrip — middleware retry (AddTrip, НЕ SetBaseTrip)
// -------------------------------------------------------------------------- //

// retryTrip — retry через middleware: пробует 3 раза при 5xx
func retryTrip[tMeta any](next reqforge.Trip[tMeta]) reqforge.Trip[tMeta] {
	return func(c *reqforge.Ctx[tMeta], req *http.Request) (*http.Response, error) {
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
//  queryEncodeRequestBody — сериализация RequestBody в query params (body = nil)
// -------------------------------------------------------------------------- //

func queryEncodeRequestBody(c *reqforge.Ctx[AppMeta], v []RequestBody, httpReq *http.Request) error {
	q := httpReq.URL.Query()
	for _, v := range v {
		q.Add("id", fmt.Sprintf("%d", v.ID))
		q.Add("payload", v.Payload)
	}
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
func mockTransportTrip[tMeta any](next reqforge.Trip[tMeta]) reqforge.Trip[tMeta] {
	return func(c *reqforge.Ctx[tMeta], req *http.Request) (*http.Response, error) {
		// Подменяем ВЕСЬ транспорт на mock (возвращает 200 OK)
		c.SetBaseTrip(&mockTransport{
			statusCode: 200,
			body:       `{"status":"mock-processed"}`,
		})
		return next(c, req)
	}
}
