package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yourusername/DistributedMediaHub/internal/api-gateway/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	cfg := config.Loader()
	err := authService.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost:"+cfg.Server.Port, opts)
	if err != nil {
		log.Fatalf("Не удалось зарегистрировать gateway: %v", err)
	}

	fmt.Println("API Gateway запущен на :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}
}
