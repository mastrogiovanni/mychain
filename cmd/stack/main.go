package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mastrogiovanni/mychain/internal/stack"
)

type Pippo struct {
	Counter int64
}

func (pippo *Pippo) Handle(direction stack.Direction, data []byte, layer *stack.Layer) {
	pippo.Counter++
	log.Println("Lower Counter", pippo.Counter, string(data))
}

func main() {

	pippo := &Pippo{Counter: 41}

	lower := stack.NewLayer(pippo.Handle)

	middle := stack.NewLayer(func(direction stack.Direction, data []byte, layer *stack.Layer) {
		switch direction {
		case stack.Up:
			log.Println("middle (going up)", direction, string(data))
			layer.SendUp([]byte(fmt.Sprintf("%s/middle", string(data))))
		case stack.Down:
			log.Println("middle (going down)", direction, string(data))
			layer.SendDown([]byte(fmt.Sprintf("%s/middle", string(data))))
		}
	})

	upper := stack.NewLayer(func(direction stack.Direction, data []byte, layer *stack.Layer) {
		log.Println("upper", direction, string(data))
		layer.SendDown(data)
	})

	upper.AttachLower(middle)
	middle.AttachLower(lower)

	for {
		lower.SendUp([]byte("Hello"))
		log.Println("---------------")
		time.Sleep(5 * time.Millisecond)
	}

}
