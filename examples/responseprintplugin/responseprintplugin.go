package main

import (
	"github.com/joncooperworks/judas"
	"log"
	"net/http/httputil"
)

type responseprintplugin struct {
	logger *log.Logger
}

// Listen pulls search queries out of HTTP exchanges
func (p *responseprintplugin) Listen(exchanges <-chan *judas.HTTPExchange) {
	for exchange := range exchanges {
		response, _ := httputil.DumpResponse(exchange.Response.Response, true)
		p.logger.Printf("\nresponse: %v\n", string(response))
	}
}

// New returns a plugin that logs searches.
func New(logger *log.Logger) (judas.Listener, error) {
	return &responseprintplugin{logger: logger}, nil
}
