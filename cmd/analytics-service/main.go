package main

import (
	"fmt"

	"github.com/yourusername/DistributedMediaHub/internal/config"
)

func main() {
	cfg := config.Loader()
	fmt.Println(cfg)

}
