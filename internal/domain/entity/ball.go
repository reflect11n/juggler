package entity

import (
	"fmt"
	"time"
)

type Ball struct {
	ID       int64
	Duration int64
}

func (b *Ball) Fly(rightHand chan<- Ball) {
	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		elapsed := int(time.Since(start).Seconds())
		fmt.Printf("Ball #%d: %d/%d seconds\n", b.ID, elapsed, b.Duration)
		if elapsed >= int(b.Duration) {
			rightHand <- *b
			return
		}
	}
}
