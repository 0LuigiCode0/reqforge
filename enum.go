package reqforge

import "net/http"

// Header — тип для имён HTTP-заголовков.
type Header = string

const (
	HeaderContentType   Header = "Content-Type"
	HeaderContentLength Header = "Content-Length"
	HeaderAuthorization Header = "Authorization"
	HeaderUserAgent     Header = "User-Agent"
)

// ContentType — тип для MIME-типов контента.
type ContentType = string

const (
	ContentTypeJson           ContentType = "application/json"
	ContentTypePdf            ContentType = "application/pdf"
	ContentTypeXml            ContentType = "application/xml"
	ContentTypeZip            ContentType = "application/zip"
	ContentTypePlain          ContentType = "text/plain"
	ContentTypeHtml           ContentType = "text/html"
	ContentTypeTextXml        ContentType = "text/xml"
	ContentTypeFormUrlencoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeForm           ContentType = "multipart/form-data"
)

// Method — тип для HTTP-методов.
type Method = string

const (
	MethodGet    Method = http.MethodGet
	MethodPost   Method = http.MethodPost
	MethodPatch  Method = http.MethodPatch
	MethodPut    Method = http.MethodPut
	MethodDelete Method = http.MethodDelete
)

type Schema = string

const (
	SchemaHttp  Schema = "http://"
	SchemaHttps Schema = "https://"
)
