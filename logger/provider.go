package logging

import (
	"github.com/firmeve/firmeve"
	"github.com/firmeve/firmeve/container"
)

type Provider struct {
	Firmeve *firmeve.Firmeve `inject:"firmeve"`
	id int
}

func (p *Provider) Register() {
	//@todo 这里需要引入config
	p.Firmeve.Bind(`logger`, Default(), container.WithShare(true))
}

func (p *Provider) Boot() {

}

func init() {
	firmeve := firmeve.New()
	firmeve.Register(`logger`, firmeve.Resolve(new(Provider)).(*Provider))
}