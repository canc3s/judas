package main

import (
	"fmt"
	"github.com/joncooperworks/judas"
	"log"
	"net/http/httputil"
	"os"
	"time"
)

type loggingplugin struct {
	logger *log.Logger
}

// Listen pulls search queries out of HTTP exchanges
func (p *loggingplugin) Listen(exchanges <-chan *judas.HTTPExchange) {
	for exchange := range exchanges {
		request, _ := httputil.DumpRequest(exchange.Request.Request, true)
		now := time.Now()
		filename := fmt.Sprintf("%d%d%d.log",now.Year(), now.Month(),now.Day())
		f, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		f.Write([]byte("\n"))
		f.Write(request)
		f.Write([]byte("\n"))
		f.Close()
	}
}

// New returns a plugin that logs searches.
func New(logger *log.Logger) (judas.Listener, error) {
	return &loggingplugin{logger: logger}, nil
}
