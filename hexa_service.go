package hrpc

import (
	"context"
	"net"

	"github.com/kamva/hexa"
	"github.com/kamva/tracer"
	"google.golang.org/grpc"
)

// HexaService implements hexa service.
type HexaService struct {
	hexa.Health
	*grpc.Server
	net.Listener
}

func NewHexaService(h hexa.Health, s *grpc.Server, l net.Listener) hexa.Service {
	return &HexaService{
		Health:   h,
		Server:   s,
		Listener: l,
	}
}

func (s *HexaService) Run() error {
	return tracer.Trace(s.Server.Serve(s.Listener))
}

func (s *HexaService) Shutdown(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

var _ hexa.Runnable = &HexaService{}
var _ hexa.Shutdownable = &HexaService{}
