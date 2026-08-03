// Package clock предоставляет источник текущего времени.
package clock

import "time"

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
type Fixed struct {
	Moment time.Time
	Step   time.Duration
}

// Now возвращает текущий момент и сдвигает его на шаг.
func (f *Fixed) Now() time.Time {
	current := f.Moment
	f.Moment = f.Moment.Add(f.Step)

	return current
}
