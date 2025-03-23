package lit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	grpcServer *grpc.Server
	addr       string
}

// NewGRPCServer creates new gRPC server
func NewGRPCServer(ctx context.Context, addr string) (GRPCServer, error) {
	opts := []GRPCOption{
		WithDefaultInterceptors(ctx),
	}

	return NewGRPCServerWithOptions(ctx, addr, opts...)
}

// NewGRPCServerWithOptions creates a new gRPC server with provided option
func NewGRPCServerWithOptions(ctx context.Context, addr string, opts ...GRPCOption) (GRPCServer, error) {
	var serverOpts []grpc.ServerOption
	for _, opt := range opts {
		opt(&serverOpts)
	}

	grpcServer := grpc.NewServer(serverOpts...)

	return GRPCServer{
		grpcServer: grpcServer,
		addr:       addr,
	}, nil
}

// Run starts gRPC server and listen for syscall
// kill (no param) default send syscall.SIGTERM
// kill -2 is syscall.SIGINT
// kill -9 is syscall. SIGKILL but can't be caught, so don't need to add it
func (srv GRPCServer) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return srv.RunWithContext(ctx)
}

// RunWithContext starts gRPC server and manages its lifecycle using given context
func (srv GRPCServer) RunWithContext(ctx context.Context) error {
	startupErr := make(chan error)

	go func() {
		fmt.Printf("gRPC server starting at %s\n", srv.addr)
		defer fmt.Println("gRPC server stopped")

		lis, err := net.Listen("tcp", srv.addr)
		if err != nil {
			startupErr <- err
			return
		}

		if err := srv.grpcServer.Serve(lis); err != nil {
			startupErr <- err
		}
	}()

	select {
	case err := <-startupErr:
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("grpc server startup error: %w", err)
		}
		return nil
	case <-ctx.Done():
		srv.stop()
		return nil
	}
}

func (srv GRPCServer) stop() {
	fmt.Printf("attempting to shutdown gracefully\n")
	defer fmt.Println("server shutdown successfully")

	srv.grpcServer.GracefulStop()
}
