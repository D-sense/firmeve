package http

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

//type MockResponseWriter struct {
//	mock.Mock
//}
//
//func (m *MockResponseWriter) Header() http.Header {
//	return http.Header{}
//}
//
//func (m *MockResponseWriter) Write(p []byte) (int, error) {
//	return len(p), nil
//}
//
//func (m *MockResponseWriter) WriteHeader(statusCode int) {
//}

func TestServerError(t *testing.T) {
	err1 := errors.New("error1")
	err2 := fmt.Errorf("error2 %w", err1)
	err3 := fmt.Errorf("error3 %w", err2)
	fmt.Println(errors.Is(err3,err1))
	hError := &Error{}
	assert.Implements(t,(*error)(nil),hError)

	herr1 := NewError(400,"err")
	var herr2 ErrorResponse
	fmt.Println(errors.As(herr1,&herr2))
	fmt.Println(herr2)

	//fmt.Println(errors.Is(err3,err1))
	//hError := Error{}
	//var anyerr IError
	//z := &anyerr
	//fmt.Println(errors.As(hError,z))
	//fmt.Println(z)
}

//func TestRecoveryString(t *testing.T) {
//	firmeve := firmeve.New()
//	//bootstrap.New(firmeve,path.RunRelative("../../testdata/config")).Boot()
//	firmeve.Bind(`logger`, logging.New(
//		config.New(path.RunRelative("../testdata/config")).Item("logging").(config.Configurator)))
//	firmeve.Boot()
//	Recovery(newContext(firmeve,new(MockResponseWriter),nil, func(c *Context) {
//		panic(`error`)
//	}))
//}
//
//func TestRecoveryError(t *testing.T) {
//	firmeve := firmeve.New()
//	//bootstrap.New(firmeve,path.RunRelative("../../testdata/config")).Boot()
//	firmeve.Bind(`logger`, logging.New(
//		config.New(path.RunRelative("../testdata/config")).Item("logging").(config.Configurator)))
//	firmeve.Boot()
//	Recovery(newContext(firmeve,new(MockResponseWriter),nil, func(c *Context) {
//		panic(errors.New(`error`))
//	}))
//}