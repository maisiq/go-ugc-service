package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	di "github.com/maisiq/go-ugc-service/internal/app"
	"github.com/maisiq/go-ugc-service/pkg/config"
	logx "github.com/maisiq/go-ugc-service/pkg/logger"
	ugcv1pb "github.com/maisiq/go-ugc-service/pkg/pb/ugcservice/v1"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()

	env, ok := os.LookupEnv("ENVIRONMENT")

	if !ok {
		env = "local"
	}
	cfgPath := fmt.Sprintf("./configs/config.%s.yaml", strings.ToLower(env))

	cfg := config.LoadConfig(cfgPath)
	app, err := di.NewApp(ctx, cfg)

	log := logx.Logger()
	log.Infof("environemnt: %v", env)

	if err != nil {
		log.Fatalf("Could not create new app: %v", err)
	}

	go runSwagger(ctx, cfg)

	err = app.Run()

	if err != nil {
		log.Fatalf("Could not run app: %v", err)
	}

}

func runSwagger(ctx context.Context, cfg *config.Config) {
	log := logx.Logger()

	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := ugcv1pb.RegisterUGCServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		fmt.Sprintf("%v:%d", cfg.Server.Host, cfg.Server.Port),
		opts,
	)

	if err != nil {
		log.Errorf("grpc-gateway: %v", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", gwMux)

	data, err := os.ReadFile("./swagger/ugcservice/v1/ugc.swagger.json")

	if err != nil {
		log.Fatalf("Cannot read the openapi schema: %v", err)
	}

	httpMux.HandleFunc(fmt.Sprintf("/%v/", cfg.Swagger.Endpoint), httpSwagger.WrapHandler)
	httpMux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	log.Info(fmt.Sprintf("Swagger UI: http://%v:%d/%v", cfg.Swagger.Host, cfg.Swagger.Port, cfg.Swagger.Endpoint))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%d", cfg.Swagger.Host, cfg.Swagger.Port), httpMux))
}
