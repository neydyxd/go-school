package l05systemdesign

import (
	"errors"
	"time"
)

// ErrOpen: автомат разомкнут, вызов отклонён без обращения к зависимости.
var ErrOpen = errors.New("l05systemdesign: circuit breaker разомкнут")

// State: состояние автомата.
type State int

const (
	Closed   State = iota // всё хорошо, вызовы идут
	Open                  // зависимость лежит, вызовы сразу отклоняются
	HalfOpen              // пробуем один вызов, чтобы понять, ожила ли
)

// String готов: "closed", "open" или "half-open" для логов и метрик.
func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Breaker: circuit breaker как конечный автомат.
type Breaker struct {
	// TODO: настройки, мьютекс, состояние, счётчик неудач, время размыкания
}

// NewBreaker создаёт автомат в состоянии Closed. threshold: сколько неудач
// подряд размыкают цепь (меньше 1 считается как 1), cooldown: сколько ждать
// в Open до пробного вызова, now: часы (time.Now в проде, фейк в тестах).
func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	panic("TODO: реализуй NewBreaker")
}

// State возвращает текущее состояние. Open, у которого cooldown уже истёк,
// показывается как HalfOpen (следующий вызов станет пробным).
func (b *Breaker) State() State {
	panic("TODO: реализуй State")
}

// Call выполняет fn по правилам автомата и возвращает её ошибку.
//
//   - Closed: fn вызывается. Неудача увеличивает счётчик неудач подряд,
//     успех его обнуляет. threshold неудач подряд: переход в Open.
//   - Open: fn не вызывается, вернуть ErrOpen. Когда с момента размыкания
//     прошло не меньше cooldown, автомат переходит в HalfOpen.
//   - HalfOpen: ровно один пробный вызов fn. Пока он идёт, остальные Call
//     получают ErrOpen, не вызывая fn. Успех пробы: Closed со сброшенным
//     счётчиком. Неудача: снова Open, cooldown отсчитывается заново.
//
// Call безопасен для конкурентного использования, а fn выполняется без
// удержания внутреннего мьютекса (она может быть долгой).
func (b *Breaker) Call(fn func() error) error {
	panic("TODO: реализуй Call")
}
