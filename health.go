package hrpc

import (
	"context"

	"github.com/kamva/hexa"
	"github.com/kamva/hexa/hlog"
	"github.com/kamva/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type healthServer struct {
}

func (h healthServer) Check(c context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h healthServer) Watch(_ *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func NewHealthServer() grpc_health_v1.HealthServer {
	return &healthServer{}
}

type grpcHealth struct {
	id   string
	addr string
}

func NewGRPCHealth(id string, addr string) hexa.Health {
	return &grpcHealth{id: id, addr: addr}
}

func (g *grpcHealth) HealthIdentifier() string {
	return "grpc_server"
}

func (g *grpcHealth) client() (grpc_health_v1.HealthClient, error) {
	c, err := grpc.Dial(g.addr, grpc.WithInsecure())
	if err != nil {
		return nil, tracer.Trace(err)
	}

	client := grpc_health_v1.NewHealthClient(c)
	return client, nil

}
func (g *grpcHealth) LivenessStatus(ctx context.Context) hexa.LivenessStatus {
	client, err := g.client()
	if err != nil {
		hlog.Error("error on creating grpc client connection", hlog.ErrStack(tracer.Trace(err)), hlog.Err(err))
		return hexa.StatusDead
	}

	res, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		hlog.Error("error on result of grpc health check call", hlog.ErrStack(tracer.Trace(err)), hlog.Err(err))
		return hexa.StatusDead
	}

	if res.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return hexa.StatusDead
	}

	return hexa.StatusAlive
}

func (g *grpcHealth) ReadinessStatus(ctx context.Context) hexa.ReadinessStatus {
	client, err := g.client()
	if err != nil {
		hlog.Error("error on creating grpc client connection", hlog.ErrStack(tracer.Trace(err)), hlog.Err(err))
		return hexa.StatusUnReady
	}

	res, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		hlog.Error("error on result of grpc health check call", hlog.ErrStack(tracer.Trace(err)), hlog.Err(err))
		return hexa.StatusUnReady
	}

	if res.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return hexa.StatusUnReady
	}

	return hexa.StatusReady
}

func (g *grpcHealth) HealthStatus(ctx context.Context) hexa.HealthStatus {
	// For now both calls to health and ready for grpc server is a call to
	// the "Check" method.
	liveness := g.LivenessStatus(ctx)
	readiness := hexa.StatusUnReady
	if liveness == hexa.StatusAlive {
		readiness = hexa.StatusReady
	}

	return hexa.HealthStatus{
		Id:    g.HealthIdentifier(),
		Alive: liveness,
		Ready: readiness,
	}
}
