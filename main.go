package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kube-svcauth-proxy/proxy"
	"log"
	"net/http"
	"sync"
)

type ProxyConf struct {
	ProxyConfigs []proxy.Conf `yaml:"proxy"`
}

func main() {
	configs, err := parseConf()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)

	for _, pc := range configs.ProxyConfigs {
		server, err := proxy.NewProxyServer(pc)
		if err != nil {
			panic(err)
		}
		go runServer(server, wg)
		wg.Add(1)
	}

	wg.Wait()
	log.Println("Good bye!")
}

func parseConf() (*ProxyConf, error) {
	configFile := ""
	flag.StringVar(&configFile, "conf", "conf.yaml", "Path to config yaml file.")
	flag.Parse()

	return readConfFile(configFile)
}

func readConfFile(filename string) (*ProxyConf, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	conf := &ProxyConf{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return conf, nil
}

func runServer(server *http.Server, wg *sync.WaitGroup) {
	log.Printf("Starting server '%s' ...\n", server.Addr)

	server.RegisterOnShutdown(func() {
		log.Printf("Shutting down server '%s'... \n", server.Addr)
		wg.Done()
	})

	log.Println("Proxy.run()", server.ListenAndServe())
}
