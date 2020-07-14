package k8stun

import (
	"sync"

	"k8s.io/client-go/rest"
)

// NewService ...
func NewService(config *rest.Config, tunnels []Tunnel) *Service {
	svc := &Service{
		wg:      &sync.WaitGroup{},
		tunnels: []*Tunnel{},
	}
	for id, tun := range tunnels {
		svc.tunnels = append(svc.tunnels, tun.Initialize(id, svc.wg, config))
	}
	return svc
}

// Service ...
type Service struct {
	wg      *sync.WaitGroup
	tunnels []*Tunnel
}

// Start all tunnels
func (s *Service) Start() {
	for _, tun := range s.tunnels {
		tun.Start()
	}
}

// Stop all tunnels
func (s *Service) Stop() {
	for _, tun := range s.tunnels {
		tun.Stop()
	}
	s.wg.Wait()
}
