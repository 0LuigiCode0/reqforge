package reqforge

import "net/http"

// Header — тип для имён HTTP-заголовков.
type Header = string

const (
	HeaderContentType   Header = "Content-Type"   // заголовок Content-Type
	HeaderContentLength Header = "Content-Length" // заголовок Content-Length
	HeaderAuthorization Header = "Authorization"  // заголовок Authorization
	HeaderUserAgent     Header = "User-Agent"     // заголовок User-Agent
)

// ContentType — тип для MIME-типов контента.
type ContentType = string

const (
	ContentTypeJson             ContentType = "application/json"
	ContentTypePdf              ContentType = "application/pdf"
	ContentTypeXml              ContentType = "application/xml"
	ContentTypeZip              ContentType = "application/zip"
	ContentTypePlain            ContentType = "text/plain"
	ContentTypeHtml             ContentType = "text/html"
	ContentTypeTextXml          ContentType = "text/xml"
	ContentTypeFormUrlencoded   ContentType = "application/x-www-form-urlencoded"
	ContentTypeForm             ContentType = "multipart/form-data"
)

// Method — тип для HTTP-методов.
type Method = string

const (
	MethodGet    Method = http.MethodGet    // GET
	MethodPost   Method = http.MethodPost   // POST
	MethodPatch  Method = http.MethodPatch  // PATCH
	MethodPut    Method = http.MethodPut    // PUT
	MethodDelete Method = http.MethodDelete // DELETE
)
