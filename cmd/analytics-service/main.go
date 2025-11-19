package main

import (
	"fmt"

	"github.com/yourusername/DistributedMediaHub/internal/analytics-service/config"
)

func main() {
	cfg := config.Loader()
	fmt.Println(cfg)

}
