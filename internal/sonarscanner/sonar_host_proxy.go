package sonarscanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

type sonarHostProxyFactory struct {
	listenAddr   string
	sonarHostUrl string
	config       *tls.Config
	log          *logrus.Entry
}

type sonarHostProxy struct {
	listenAddr string
	log        *logrus.Entry
	proxy      *httputil.ReverseProxy
}

func (f *sonarHostProxyFactory) new() (*sonarHostProxy, error) {
	target, err := url.Parse(f.sonarHostUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sonar host url: %s", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		TLSClientConfig: f.config,
	}

	return &sonarHostProxy{
		listenAddr: f.listenAddr,
		log:        f.log,
		proxy:      proxy,
	}, nil
}

func (p *sonarHostProxy) serveWithContext(ctx context.Context) error {
	p.log.Infof("Starting reverse proxy on %s ...", p.listenAddr)

	server := &http.Server{
		Addr: p.listenAddr,
		Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			p.log.Debugf("Proxying request %s %s", req.Method, req.URL)
			p.proxy.ServeHTTP(res, req)
		}),
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		p.log.Info("Stopping the reverse proxy ...")
		if err := server.Close(); err != nil {
			p.log.Warnf("Failed to stop the reverse proxy: %s", err)
		} else {
			p.log.Info("Reverse proxy stopped")
		}

		return nil
	}
}
