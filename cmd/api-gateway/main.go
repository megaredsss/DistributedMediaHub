package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	r := chi.NewRouter()
	opts := []grpc.DialOption{grpc.WithTransportCredentials((insecure.NewCredentials()))}
	http.ListenAndServe(":8081", r)
}
