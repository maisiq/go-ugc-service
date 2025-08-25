package app

import (
	"context"
	"net"
	"strconv"

	"github.com/maisiq/go-ugc-service/internal/closer"
	"github.com/maisiq/go-ugc-service/internal/server"
	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/maisiq/go-ugc-service/pkg/logger"
	ugcv1pb "github.com/maisiq/go-ugc-service/pkg/pb/ugcservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg             *config.Config
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	a := &App{cfg: cfg}

	err := a.initDeps(ctx)
	if err != nil {
		log := logger.Logger()
		log.Fatal("Failed to init deps")
		return nil, err
	}
	return a, nil
}

func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	return a.runGRPCServer()

}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initLogger,
		a.initServiceProvider,
		a.initGRPCServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	logger.InitLogger(a.cfg.App.Debug)
	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider(a.cfg)
	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(server.ValidateInterceptor),
	)

	reflection.Register(a.grpcServer)

	ugcv1pb.RegisterUGCServiceServer(a.grpcServer, a.serviceProvider.UGCServiceServer(ctx))
	return nil
}

func (a *App) runGRPCServer() error {
	log := a.serviceProvider.Logger()
	log.Infof("GRPC server is running on %v:%v", a.cfg.Server.Host, a.cfg.Server.Port)

	lis, err := net.Listen(
		"tcp",
		net.JoinHostPort(a.cfg.Server.Host, strconv.Itoa(a.cfg.Server.Port)),
	)

	if err != nil {
		log.Fatal(err)
	}

	closer.Add(func() error {
		log.Info("GRPC: Graceful shutdown")
		a.grpcServer.GracefulStop()
		return nil
	})

	err = a.grpcServer.Serve(lis)

	if err != nil {
		log.Errorf("Failed to serve server: %v", err)
	}

	return nil
}
