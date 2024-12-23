package stat

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

// MsgStat 统计消息执行时间 groutine safe
type MsgStat[K comparable] struct {
	mu   sync.RWMutex   //
	msgs map[K]*stat[K] //
}

// NewStat 创建Stat管理
func NewStat[K comparable]() *MsgStat[K] {
	return &MsgStat[K]{
		msgs: make(map[K]*stat[K]),
	}
}

func (m *MsgStat[K]) get(id K, cost int64) *stat[K] {
	m.mu.RLock()
	s := m.msgs[id]
	m.mu.RUnlock()
	if s != nil {
		return s
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if s = m.msgs[id]; s != nil {
		return s
	}

	s = &stat[K]{
		min: cost,
		max: cost,
	}
	m.msgs[id] = s
	return s
}

// Add 不统计id为nil的消息
func (m *MsgStat[K]) Add(id K, cost int64) {
	stat := m.get(id, cost)
	stat.add(cost)
}

func (m *MsgStat[K]) statistic() []Statistic {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs := make([]Statistic, 0, len(m.msgs))
	for id, s := range m.msgs {
		msgs = append(msgs, s.statistic(id))
	}
	return msgs
}

func (*MsgStat[K]) stringStatistic(
	buf *bytes.Buffer, stats []Statistic, less func(i, j int) bool, v func(*Statistic) any,
) string {
	buf.Reset()
	sort.Slice(stats, less)
	for i := range stats {
		if i > 0 {
			_, _ = buf.WriteString(", ")
		}
		switch v := v(&stats[i]); v.(type) {
		case float64:
			_, _ = buf.WriteString(fmt.Sprintf("%v:%.2f", stats[i].ID, v))
		default:
			_, _ = buf.WriteString(fmt.Sprintf("%v:%v", stats[i].ID, v))
		}
	}
	return buf.String()
}

// Statistic 统计信息, 一般用于 print
func (m *MsgStat[K]) Statistic() map[string]string {
	stats := m.statistic()
	// pct
	var total int64
	for i := range stats {
		total += stats[i].Total
	}
	for i := range stats {
		stats[i].Pct = float64(stats[i].Total) / float64(total)
	}

	var buf bytes.Buffer
	return map[string]string{
		"total": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Total > stats[j].Total
			},
			func(s *Statistic) any {
				return s.Total
			},
		),
		"cnt": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Cnt > stats[j].Cnt
			},
			func(s *Statistic) any {
				return s.Cnt
			},
		),
		"min": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Min > stats[j].Min
			},
			func(s *Statistic) any {
				return s.Min
			},
		),
		"max": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Max > stats[j].Max
			},
			func(s *Statistic) any {
				return s.Max
			},
		),
		"pct": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Pct > stats[j].Pct
			},
			func(s *Statistic) any {
				return s.Pct
			},
		),
		"avg": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Avg > stats[j].Avg
			},
			func(s *Statistic) any {
				return s.Avg
			},
		),
		"var": m.stringStatistic(
			&buf, stats,
			func(i, j int) bool {
				return stats[i].Var > stats[j].Var
			},
			func(s *Statistic) any {
				return s.Var
			},
		),
	}
}
