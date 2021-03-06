package binding

import (
	"fmt"
	"github.com/firmeve/firmeve/kernel/contract"
)

var (
	httpBindingType = map[string]contract.Binding{
		contract.HttpMimeJson:          JSON,
		contract.HttpMimeForm:          Form,
		contract.HttpMimeMultipartForm: MultipartForm,
	}
)

func Bind(protocol contract.Protocol, v interface{}) error {
	if p, ok := protocol.(contract.HttpProtocol); ok {
		contentType := p.ContentType()

		if b, ok := httpBindingType[contentType]; ok {
			return b.Protocol(protocol, v)
		}

		return fmt.Errorf("non-existent type %s", contentType)
	}

	return nil
}
