package main

import (
	"github.com/joncooperworks/judas"
	"log"
	"net/http/httputil"
)

type requestprintplugin struct {
	logger *log.Logger
}

// Listen pulls search queries out of HTTP exchanges
func (p *requestprintplugin) Listen(exchanges <-chan *judas.HTTPExchange) {
	for exchange := range exchanges {
		request, _ := httputil.DumpRequest(exchange.Request.Request, true)
		p.logger.Printf("\nrequest: %v\n", string(request))
	}
}

// New returns a plugin that logs searches.
func New(logger *log.Logger) (judas.Listener, error) {
	return &requestprintplugin{logger: logger}, nil
}
