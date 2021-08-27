package proxy

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type handler struct {
	proxy           *httputil.ReverseProxy
	conf            *Conf
	authorizedToken string
}

type Conf struct {
	Listen                string `yaml:"listen"`
	Upstream              string `yaml:"upstream"`
	K8sAuthcheckNamespace string `yaml:"namespace"`
	K8sAuthcheckService   string `yaml:"service"`
}

func (ph *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := ph.authorizeAccess(r.Header.Get("Authorization")); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	} else {
		ph.proxy.ServeHTTP(w, r)
	}
}

func (ph *handler) authorizeAccess(token string) error {
	if token != "" {
		// let through token that was authorized before
		if token == ph.authorizedToken {
			return nil
		}

		// if not cached authorized token, let's check kubernetes service get access
		err := ph.authorizeK8s(token)
		if err != nil {
			log.Printf("Unauthorized access at '%s'\n", ph.conf.Listen)
			log.Println(err)
			return fmt.Errorf("unauthorized access")
		} else {
			log.Printf("Token authorized and cached for '%s'\n", ph.conf.Listen)
			// remember successfully authorized token
			ph.authorizedToken = token
			return nil
		}
	} else {
		log.Printf("Unauthorized access without token header at '%s'\n", ph.conf.Listen)
		return fmt.Errorf("no authorization token header found, unauthorized access")
	}
}

func (ph *handler) authorizeK8s(token string) error {
	client, err := newK8sClient(token)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Services(ph.conf.K8sAuthcheckNamespace).Get(context.TODO(), ph.conf.K8sAuthcheckService, v1.GetOptions{})
	return err
}

func newK8sClient(token string) (*kubernetes.Clientset, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("failed to create k8s client to authorize token")
	}

	c := &rest.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		APIPath:         "/",
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		BearerToken:     stripToken(token),
	}

	client, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func stripToken(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token[len("bearer "):]
	}
	return token
}

func NewProxyServer(pc Conf) (*http.Server, error) {
	log.Printf("Creating new reverse proxy server '%+v'", pc)
	upstream, err := url.Parse(pc.Upstream)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)
	server := &http.Server{
		Addr:    pc.Listen,
		Handler: &handler{proxy: proxy, conf: &pc},
	}

	return server, nil
}
