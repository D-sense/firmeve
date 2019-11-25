package http

import (
	"encoding/json"
	"fmt"
	"github.com/firmeve/firmeve"
	resource2 "github.com/firmeve/firmeve/converter/resource"
	"github.com/firmeve/firmeve/converter/serializer"
	"github.com/firmeve/firmeve/support/strings"
	"github.com/go-playground/form/v4"
	"mime/multipart"
	"net/http"
	strings2 "strings"
	"time"
)

type (
	HandlerFunc func(c *Context)
	Params map[string]string
	entity struct {
		Key   string
		Value interface{}
	}
	Context struct {
		Firmeve        *firmeve.Firmeve `inject:"firmeve"`
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		handlers       []HandlerFunc
		entities       map[string]*entity
		index          int
		Params         Params
		route          *Route
		startTime      time.Time
	}
)

var (
	formDecoder = form.NewDecoder()
)

func newContext(firmeve *firmeve.Firmeve, writer http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) *Context {
	return &Context{
		Firmeve:        firmeve,
		Request:        r,
		ResponseWriter: writer,
		handlers:       handlers,
		entities:       make(map[string]*entity, 0),
		index:          0,
		Params:         make(Params, 0),
		startTime:      time.Now(),
	}
}

func (c *Context) SetParams(params Params) *Context {
	c.Params = params
	return c
}

func (c *Context) SetRoute(route *Route) *Context {
	c.route = route
	return c
}

func (c *Context) AddEntity(key string, value interface{}) *Context {
	c.entities[key] = &entity{
		Key:   key,
		Value: value,
	}
	return c
}

func (c *Context) Entity(key string) *entity {
	if v, ok := c.entities[key]; ok {
		return v
	}

	return nil
}

func (c *Context) EntityValue(key string) interface{} {
	if v, ok := c.entities[key]; ok {
		return v.Value
	}

	return nil
}

func (c *Context) FormDecode(v interface{}) interface{} {
	if c.Request.Form == nil {
		c.Request.ParseMultipartForm(32 << 20)
	}

	if err := formDecoder.Decode(v, c.Request.Form); err != nil {
		panic(err)
	}

	return v
}

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	f, fh, err := c.Request.FormFile(key)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}

func (c *Context) Abort(code int, message string) {
	c.AbortWithError(code, message, nil)
}

func (c *Context) AbortWithError(code int, message string, err error) {
	NewErrorWithError(code, message, err).Response(c)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Query(key string) interface{} {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Form(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) Status(code int) *Context {
	c.ResponseWriter.WriteHeader(code)
	return c
}

func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) *Context {
	c.ResponseWriter.Header().Set(key, value)
	return c
}

func (c *Context) Post(key string) string {
	return c.Request.Form.Get(key)
}

func (c *Context) Write(bytes []byte) *Context {
	_, err := c.ResponseWriter.Write(bytes)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Context) NoContent() *Context {
	c.ResponseWriter.WriteHeader(204)
	return c
}

func (c *Context) Created() *Context {
	c.ResponseWriter.WriteHeader(201)
	return c
}

func (c *Context) String(content string) *Context {
	c.Write([]byte(content))
	return c
}

func (c *Context) IsJSON() bool {
	accept := c.Header(strings.UcFirst(`Accept`))
	accepts := strings2.Split(accept, `,`)
	for _, item := range accepts {
		if item == `application/json` || item == `application+json` {
			return true
		}
	}

	return false
}

func (c *Context) JSON(content interface{}) *Context {
	c.SetHeader(`Content-Type`, `application/json`)

	str, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}
	c.Write(str)

	return c
}

func (c *Context) Data(content interface{}) *Context {
	return c.JSON(serializer.NewData(content).Resolve())
}

func (c *Context) Item(resource interface{}, option *resource2.Option) *Context {
	return c.Data(resource2.NewItem(resource, option))
}

func (c *Context) Collection(resource interface{}, option *resource2.Option) *Context {
	return c.Data(resource2.NewCollection(resource, option))
}

// JSONP serializes the given struct as JSON into the responseWriter body.
// It add padding to responseWriter body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
//func (c *Context) JSONP(code int, obj interface{}) {
//	callback := c.DefaultQuery("callback", "")
//	if callback == "" {
//		c.Render(code, render.JSON{Data: obj})
//		return
//	}
//	c.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
//}

func (c *Context) Redirect(location string, code int) {
	http.Redirect(c.ResponseWriter, c.Request, location, code)
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.ResponseWriter, c.Request, filepath)
}

func (c *Context) FileAttachment(filepath, filename string) {
	c.ResponseWriter.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeFile(c.ResponseWriter, c.Request, filepath)
}

func (c *Context) Flush() *Context {
	c.ResponseWriter.(http.Flusher).Flush()
	return c
}

func (c *Context) Next() {
	if c.index < len(c.handlers) {
		c.index++
		c.handlers[c.index-1](c)
	}
}
