package main

import (
	"net/url"
	"strconv"
	"strings"
	"fmt"
	"log"

	"github.com/coreos/go-etcd/etcd"
)

type HAProxyRegistry struct {
	client *etcd.Client
	domain string
	scope  string
}

func NewHAProxyRegistry(uri *url.URL) ServiceRegistry {
	urls := make([]string, 0)
	if uri.Host != "" {
		urls = append(urls, "http://"+uri.Host)
	}

	parts := strings.SplitN(uri.Path, "/", 3)

	return &HAProxyRegistry{client: etcd.NewClient(urls), domain: parts[2], scope: parts[1]}
}

func (r *HAProxyRegistry) Register(service *Service) error {
	if service.pp.ExposedPort == "" {
		log.Println("Skip proxy backend registration for host container", service.ID)
		return nil
	}

	port := strconv.Itoa(service.Port)
	record := `{"host":"` + service.IP + `","port":` + port + `}`
	log.Println("HAProxy register host " + service.IP + "," + " port:" + port + " proxy="+r.etcdPath(service))

	_, err := r.client.Set(r.etcdPath(service), record, uint64(service.TTL))

	if (err != nil) {
		log.Println("haproxy: unable to register", service, err)
	}

	return err
}

func (r *HAProxyRegistry) Deregister(service *Service) error {
	log.Println("HAProxy DeRegister host " + service.IP + "," + " port:" + strconv.Itoa(service.Port) + " proxy="+r.etcdPath(service))
	_, err := r.client.Delete(r.etcdPath(service), false)
	return err
}

func (r *HAProxyRegistry) Refresh(service *Service) error {
	return r.Register(service)
}

func (r *HAProxyRegistry) etcdPath(service *Service) string {
	return fmt.Sprintf("%s/proxy/%s/%s.%s/%s",
		               r.scope, service.pp.ExposedPort, service.Name, r.domain, service.ID)
}
