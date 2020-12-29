package sonarscanner

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

type sonarHostProxy struct {
	listenAddress string
	proxy         *httputil.ReverseProxy
	logger        *logrus.Entry
}

func newSonarHostProxy(
	logger *logrus.Entry,
	listenAddress string,
	destination string,
) (*sonarHostProxy, error) {
	target, err := url.Parse(destination)
	if err != nil {
		return nil, err
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &sonarHostProxy{
		logger:        logger,
		listenAddress: listenAddress,
		proxy:         reverseProxy,
	}, nil
}

func (proxy *sonarHostProxy) runWithContext(ctx context.Context) error {
	handlerFunc := func(res http.ResponseWriter, req *http.Request) {
		proxy.handleRequest(res, req)
	}
	server := &http.Server{
		Addr:    proxy.listenAddress,
		Handler: http.HandlerFunc(handlerFunc),
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	defer server.Close()

	select {
	case <-ctx.Done():
		return nil
	}
}

func (proxy *sonarHostProxy) handleRequest(
	res http.ResponseWriter,
	req *http.Request,
) {
	proxy.logger.Debugf("Redirecting %s %s.", req.Method, req.RequestURI)
	proxy.proxy.ServeHTTP(res, req)
}
