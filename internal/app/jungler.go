package app

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/reflect11n/juggler/internal/domain/entity"
)

type Jungler struct {
	ctx          context.Context
	BallCounter  atomic.Int64
	InFlight     atomic.Int64
	LeftHand     chan entity.Ball
	RightHand    chan entity.Ball
	StopJuggling <-chan struct{}
	End          atomic.Bool
	WG           *sync.WaitGroup
	nBalls       int
}

func NewJungler(ctx context.Context, nBalls int) *Jungler {
	return &Jungler{
		ctx:          ctx,
		LeftHand:     make(chan entity.Ball),
		RightHand:    make(chan entity.Ball),
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
		ball := j.NewBall()
		j.LeftHand <- ball
	}
}

// кол-во в полете, в руках, id мячей и их состояние
func (j *Jungler) monitorStatus() {
	fmt.Printf("Статус: В воздухе %d, В руках %d\n\n", j.InFlight.Load(), int64(j.nBalls)-j.InFlight.Load())
}
func (j *Jungler) Init() {
	go j.LeftHandListener()
	go j.RightHandListener()
	go j.monitorStatus()
}
func (j *Jungler) LeftHandListener() {
	for ball := range j.LeftHand {
		j.monitorStatus()

		j.InFlight.Add(1)
		j.WG.Add(1)
		go ball.Fly(j.RightHand)
	}
}

func (j *Jungler) RightHandListener() {
	for ball := range j.RightHand {
		j.InFlight.Add(-1)
		j.WG.Done()
		if !j.End.Load() {
			j.LeftHand <- ball
		}
	}
}

func (j *Jungler) HandleShutDown() {
	<-j.StopJuggling
	j.End.Store(true)
	close(j.LeftHand)
	close(j.RightHand)
}
