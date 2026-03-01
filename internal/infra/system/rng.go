package system

import (
	"math/rand"
	"time"
)

type MathRandom struct {
	rnd *rand.Rand
}

func NewMathRandom(seed int64) *MathRandom {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &MathRandom{rnd: rand.New(rand.NewSource(seed))}
}

func (r *MathRandom) Intn(n int) int {
	return r.rnd.Intn(n)
}
