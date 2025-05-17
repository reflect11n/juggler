package app

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reflect11n/juggler/internal/domain/entity"
)

type Jungler struct {
	ctx          context.Context
	BallCounter  atomic.Int64
	InFlight     atomic.Int64
	LeftHand     chan entity.Ball
	RightHand    chan entity.Ball
	StopJuggling <-chan struct{}
	balls        sync.Map
	End          atomic.Bool
	WG           *sync.WaitGroup
	nBalls       int
}

func NewJungler(ctx context.Context, nBalls int) *Jungler {
	return &Jungler{
		ctx:          ctx,
		LeftHand:     make(chan entity.Ball),
		RightHand:    make(chan entity.Ball),
		balls:        sync.Map{},
		WG:           &sync.WaitGroup{},
		StopJuggling: ctx.Done(),
		nBalls:       nBalls,
	}
}

func (j *Jungler) NewBall() entity.Ball {
	id := j.BallCounter.Add(1)
	return entity.Ball{
		ID:       id,
		Duration: int64(5 + rand.Intn(5)),
	}
}

func (j *Jungler) StartJungling() {
	for i := 0; i < j.nBalls; i++ {
		select {
		case <-j.StopJuggling:
			return
		default:
		}

		ball := j.NewBall()
		j.LeftHand <- ball

		time.Sleep(time.Second)
	}
}

func (j *Jungler) monitorStatus() {
	inFlight := j.InFlight.Load()
	inHands := int64(j.nBalls) - inFlight

	fmt.Printf("Статус: В воздухе %d, В руках %d\n", inFlight, inHands)

	for id := int64(1); id <= int64(j.nBalls); id++ {
		if _, ok := j.balls.Load(id); ok {
			fmt.Printf("  Ball#%d: в полёте\n", id)
		} else {
			fmt.Printf("  Ball#%d: в руках\n", id)
		}
	}
	fmt.Println()
}

func (j *Jungler) Init() {
	go j.LeftHandProcessor()
	go j.RightHandProcessor()
}

func (j *Jungler) LeftHandProcessor() {
	defer close(j.RightHand)
	for ball := range j.LeftHand {
		j.monitorStatus()
		j.InFlight.Add(1)
		j.balls.Store(ball.ID, ball)
		j.WG.Add(1)
		go ball.Fly(j.RightHand)
	}
}

func (j *Jungler) RightHandProcessor() {
	defer close(j.LeftHand)
	for ball := range j.RightHand {
		j.InFlight.Add(-1)
		j.balls.Delete(ball.ID)
		j.WG.Done()
		if !j.End.Load() {
			j.LeftHand <- ball
		}
	}
}

func (j *Jungler) HandleShutDown() {
	<-j.StopJuggling
	fmt.Println("Время вышло. Новых мячей не будет")
	j.End.Store(true)
	j.WG.Wait()
	fmt.Println("Все мячи упали.")
}
