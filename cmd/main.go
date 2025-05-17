package main

import (
	"context"
	"fmt"
	"time"

	"github.com/reflect11n/juggler/internal/app"
)

const (
	tMin   = 3
	nBalls = 3
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*tMin)
	defer cancel()

	j := app.NewJungler(ctx, nBalls)

	j.Init()
	fmt.Println("Жонглер активирован...")

	fmt.Println("Начинаем подбрасывать мячи...")
	go j.StartJungling()

	j.HandleShutDown()
}
