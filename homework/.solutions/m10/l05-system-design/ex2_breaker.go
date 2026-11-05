package l05systemdesign

import (
	"errors"
	"sync"
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
	threshold int
	cooldown  time.Duration
	now       func() time.Time

	mu       sync.Mutex
	state    State
	failures int       // подряд, в Closed
	openedAt time.Time // когда перешли в Open
	probing  bool      // в HalfOpen пробный вызов уже идёт
}

// NewBreaker создаёт автомат в состоянии Closed. threshold: сколько неудач
// подряд размыкают цепь (меньше 1 считается как 1), cooldown: сколько ждать
// в Open до пробного вызова, now: часы (time.Now в проде, фейк в тестах).
func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	return &Breaker{threshold: max(threshold, 1), cooldown: cooldown, now: now}
}

// State возвращает текущее состояние. Open, у которого cooldown уже истёк,
// показывается как HalfOpen (следующий вызов станет пробным).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == Open && b.now().Sub(b.openedAt) >= b.cooldown {
		return HalfOpen
	}
	return b.state
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
	b.mu.Lock()
	if b.state == Open && b.now().Sub(b.openedAt) >= b.cooldown {
		b.state = HalfOpen
		b.probing = false
	}
	if b.state == Open || b.state == HalfOpen && b.probing {
		b.mu.Unlock()
		return ErrOpen
	}
	probe := b.state == HalfOpen
	b.probing = probe
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()
	switch {
	case probe && err == nil:
		b.state, b.failures, b.probing = Closed, 0, false
	case probe:
		b.state, b.openedAt, b.probing = Open, b.now(), false
	case b.state != Closed:
		// Вызов начался в Closed, но пока он шёл, цепь уже разомкнули
		// другие вызовы: его результат ничего не меняет.
	case err == nil:
		b.failures = 0
	default:
		b.failures++
		if b.failures >= b.threshold {
			b.state, b.openedAt, b.failures = Open, b.now(), 0
		}
	}
	return err
}
