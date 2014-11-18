package main

import (
	"net/url"
	"strconv"
	"strings"
    "fmt"

	"github.com/coreos/go-etcd/etcd"
)

type Skydns2Registry struct {
	client *etcd.Client
	path   string
	domain string
	scope  string
}

func NewSkydns2Registry(uri *url.URL) ServiceRegistry {
	urls := make([]string, 0)
	if uri.Host != "" {
		urls = append(urls, "http://"+uri.Host)
	}

	parts := strings.SplitN(uri.Path, "/", 2)

	return &Skydns2Registry{client: etcd.NewClient(urls), path: domainPath(parts[3]), domain: parts[3], scope: parts[2]}
}

func (r *Skydns2Registry) Register(service *Service) error {
	port := strconv.Itoa(service.Port)
	record := `{"host":"` + service.IP + `","port":` + port + `}`
	_, err := r.client.Set(r.skydnsPath(service), record, uint64(service.TTL))
	if (err == nil) {
		_, err2 := r.client.Set(r.proxyPath(service), record, uint64(service.TTL))
		return err2
	}
	return err
}

func (r *Skydns2Registry) Deregister(service *Service) error {
	_, err := r.client.Delete(r.skydnsPath(service), false)
	if (err == nil) {
		_, err2 := r.client.Delete(r.proxyPath(service), false)
		return err2
	}
	return err
}

func (r *Skydns2Registry) Refresh(service *Service) error {
	return r.Register(service)
}

func (r *Skydns2Registry) skydnsPath(service *Service) string {
	return r.scope + "/skydns/" + r.path + "/" +
		   reversePath(service.Name) + "/" + service.pp.ExposedPort + "/" +
		   strings.Replace(service.ID, ":", "-", -1)
}

func (r *Skydns2Registry) proxyPath(service *Service) string {
	return fmt.Sprintf("%s/proxy/%s.%s/%s", r.scope, service.pp.ExposedPort, service.Name, r.domain, strings.Replace(service.ID, ":", "-", -1))
}

func reversePath(domain string) string {
	components := strings.Split(domain, ".")
	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}
	return strings.Join(components, "/")
}

func domainPath(domain string) string {
	return "/skydns/" + reversePath(domain)
}
