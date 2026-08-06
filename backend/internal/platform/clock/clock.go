// Package clock предоставляет источник текущего времени.
package clock

import (
	"sync"
	"time"
)

// Clock возвращает текущее время. Отдельный интерфейс нужен, чтобы тесты
// проверяли метки времени попытки детерминированно.
type Clock interface {
	Now() time.Time
}

// System использует системное время в UTC.
type System struct{}

// Now возвращает текущее системное время в UTC.
func (System) Now() time.Time {
	return time.Now().UTC()
}

// Fixed выдаёт заранее заданный момент и сдвигает его на Step после каждого
// обращения, поэтому в тесте виден понятный порядок событий.
//
// Мьютекс нужен, потому что тесты конкурентности вызывают Now из нескольких
// горутин: без него детектор гонок находит запись в Moment.
type Fixed struct {
	Moment time.Time
	Step   time.Duration

	mu sync.Mutex
}

// Now возвращает текущий момент и сдвигает его на шаг.
func (f *Fixed) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	current := f.Moment
	f.Moment = f.Moment.Add(f.Step)

	return current
}
