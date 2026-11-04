package l03testabledesign

import (
	"errors"
	"fmt"
	"testing"
)

func TestCheckRemindAcceptsCorrect(t *testing.T) {
	// Правильные реализации подаются обёртками, а не самой Remind: сравнить
	// функцию с Remind по адресу и так «угадать» эталон не выйдет.
	correct := []struct {
		name string
		fn   RemindFunc
	}{
		{"обёртка над Remind", func(ms []Member, m Mailer) (int, error) { return Remind(ms, m) }},
		{"независимая реализация", okRule().run},
	}
	for _, c := range correct {
		t.Run(c.name, func(t *testing.T) {
			if err := CheckRemind(c.fn); err != nil {
				t.Fatalf("CheckRemind(правильная функция) = %v, want nil: ложная тревога", err)
			}
		})
	}
}

func TestCheckRemindCatchesMutants(t *testing.T) {
	for _, m := range remindMutants {
		t.Run(m.name, func(t *testing.T) {
			if !remindCaught(m.fn) {
				t.Errorf("мутант выжил: CheckRemind вернул nil")
			}
		})
	}
}

// remindCaught: CheckRemind нашёл нарушение. Мутанты не паникуют,
// поэтому паника означает ошибку в самой проверке и роняет тест.
func remindCaught(fn RemindFunc) bool {
	return CheckRemind(fn) != nil
}

// remindRule это Remind с подменяемыми правилами: из неё собираются мутанты.
type remindRule struct {
	skipInactive bool
	minDays      int
	maxDays      int
	subject      func(Member) string
	stopOnErr    bool
	countFailed  bool
	dropErrs     bool
	sendTwice    bool
}

func (r remindRule) run(members []Member, m Mailer) (int, error) {
	sent := 0
	var errs []error
	for _, mb := range members {
		if (r.skipInactive && !mb.Active) || mb.DaysLeft < r.minDays || mb.DaysLeft > r.maxDays {
			continue
		}
		times := 1
		if r.sendTwice {
			times = 2
		}
		for range times {
			if err := m.Send(mb.Email, r.subject(mb)); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", mb.Email, err))
				if r.stopOnErr {
					return sent, errors.Join(errs...)
				}
				if !r.countFailed {
					continue
				}
			}
			sent++
		}
	}
	if r.dropErrs {
		return sent, nil
	}
	return sent, errors.Join(errs...)
}

func okRule() remindRule {
	return remindRule{
		skipInactive: true, minDays: 0, maxDays: 3,
		subject: func(mb Member) string {
			return fmt.Sprintf("Подписка заканчивается через %d дн.", mb.DaysLeft)
		},
	}
}

func mutant(change func(*remindRule)) RemindFunc {
	r := okRule()
	change(&r)
	return r.run
}

var remindMutants = []struct {
	name string
	fn   RemindFunc
}{
	{"пишет неактивным", mutant(func(r *remindRule) { r.skipInactive = false })},
	{"пропускает DaysLeft == 3", mutant(func(r *remindRule) { r.maxDays = 2 })},
	{"пишет при DaysLeft == 4", mutant(func(r *remindRule) { r.maxDays = 4 })},
	{"пропускает DaysLeft == 0", mutant(func(r *remindRule) { r.minDays = 1 })},
	{"пишет просроченным", mutant(func(r *remindRule) { r.minDays = -1 })},
	{"тема без числа дней", mutant(func(r *remindRule) {
		r.subject = func(Member) string { return "Подписка заканчивается" }
	})},
	{"останавливается на первой ошибке", mutant(func(r *remindRule) { r.stopOnErr = true })},
	{"считает неудачные отправки", mutant(func(r *remindRule) { r.countFailed = true })},
	{"глотает ошибки отправки", mutant(func(r *remindRule) { r.dropErrs = true })},
	{"шлёт письмо дважды", mutant(func(r *remindRule) { r.sendTwice = true })},
	{"теряет ошибку отправителя", func(members []Member, m Mailer) (int, error) {
		n, err := Remind(members, m)
		if err != nil {
			return n, errors.New("send failed") // без %w: errors.Is не найдёт исходную
		}
		return n, nil
	}},
}
