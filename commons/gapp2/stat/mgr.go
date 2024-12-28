package stat

import (
	"sync"
)

type Statistic struct {
	ID    any     //
	Total int64   //
	Cnt   int64   //
	Min   int64   //
	Max   int64   //
	Pct   float64 //
	Avg   float64 //
	Var   float64 //
}

type stat[K comparable] struct {
	mu       sync.Mutex //
	total    int64      //
	cnt      int64      //
	min      int64      //
	max      int64      //
	avg      float64    //
	variance float64    //
}

func (s *stat[K]) add(x int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.total += x
	s.cnt++
	s.max = max(s.max, x)
	s.min = min(s.min, x)

	avg := float64(s.total) / float64(s.cnt)
	s.variance += (float64(x) - avg) * (float64(x) - s.avg)
	s.avg = avg
}

func (s *stat[K]) statistic(id K) Statistic {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Statistic{
		ID:    id,
		Total: s.total,
		Cnt:   s.cnt,
		Min:   s.min,
		Max:   s.max,
		Avg:   s.avg,
		Var:   s.variance,
	}
}
