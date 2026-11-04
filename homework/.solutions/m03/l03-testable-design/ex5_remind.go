package l03testabledesign

import (
	"errors"
	"fmt"
)

// Member это подписчик сервиса.
type Member struct {
	Email    string
	Active   bool
	DaysLeft int // сколько дней осталось до конца подписки
}

// Mailer отправляет письма. В проде это SMTP или внешний API.
type Mailer interface {
	Send(to, subject string) error
}

// Remind рассылает напоминания о скором конце подписки.
// Это готовая и правильная функция, менять её не нужно.
//
// Правила:
//   - письмо получает каждый активный подписчик с 0 <= DaysLeft <= 3, ровно одно;
//   - тема письма: fmt.Sprintf("Подписка заканчивается через %d дн.", DaysLeft);
//   - ошибка отправки одному не останавливает рассылку остальным;
//   - результат: число успешно отправленных писем и все ошибки отправки,
//     собранные через errors.Join (nil, если ошибок не было).
func Remind(members []Member, m Mailer) (int, error) {
	sent := 0
	var errs []error
	for _, mb := range members {
		if !mb.Active || mb.DaysLeft < 0 || mb.DaysLeft > 3 {
			continue
		}
		subject := fmt.Sprintf("Подписка заканчивается через %d дн.", mb.DaysLeft)
		if err := m.Send(mb.Email, subject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", mb.Email, err))
			continue
		}
		sent++
	}
	return sent, errors.Join(errs...)
}

// RemindFunc это сигнатура Remind: так в проверку можно подать и мутанта.
type RemindFunc func(members []Member, m Mailer) (int, error)

// CheckRemind это тест в виде функции, который пишешь ты. Он вызывает remind
// на своих данных со своим шпионом-Mailer и возвращает ошибку с описанием
// первого нарушенного правила Remind или nil, если всё верно.
//
// Что понадобится:
//   - тип-шпион, реализующий Mailer: запоминает каждый вызов Send (кому и с
//     какой темой) и умеет возвращать ошибку для заданного адреса;
//   - набор подписчиков, где есть все интересные случаи: неактивный,
//     границы DaysLeft (-1, 0, 3, 4), несколько получателей;
//   - сверка: кто получил письма, сколько раз, с какой темой, что вернулось
//     в count и error (errors.Is поможет найти ошибку шпиона внутри Join).
//
// Тест прогонит CheckRemind против Remind (должно быть nil) и против
// мутантов из ex5_remind_test.go (каждого надо поймать).
func CheckRemind(remind RemindFunc) error {
	errBox := errors.New("mailbox full")
	spy := &spyMailer{failFor: "fail@x.ru", err: errBox}
	members := []Member{
		{Email: "fail@x.ru", Active: true, DaysLeft: 1},
		{Email: "zero@x.ru", Active: true, DaysLeft: 0},
		{Email: "three@x.ru", Active: true, DaysLeft: 3},
		{Email: "four@x.ru", Active: true, DaysLeft: 4},
		{Email: "expired@x.ru", Active: true, DaysLeft: -1},
		{Email: "inactive@x.ru", Active: false, DaysLeft: 2},
	}
	n, err := remind(members, spy)

	want := map[string]string{
		"fail@x.ru":  "Подписка заканчивается через 1 дн.",
		"zero@x.ru":  "Подписка заканчивается через 0 дн.",
		"three@x.ru": "Подписка заканчивается через 3 дн.",
	}
	got := map[string]string{}
	for _, c := range spy.calls {
		if _, dup := got[c.to]; dup {
			return fmt.Errorf("письмо для %s отправлено больше одного раза", c.to)
		}
		got[c.to] = c.subject
	}
	for to, subj := range want {
		if got[to] != subj {
			return fmt.Errorf("для %s тема %q, want %q", to, got[to], subj)
		}
	}
	if len(got) != len(want) {
		return fmt.Errorf("письма получили %v, want только %v", got, want)
	}
	if n != 2 {
		return fmt.Errorf("remind вернул count %d, want 2 (одна отправка из трёх упала)", n)
	}
	if !errors.Is(err, errBox) {
		return fmt.Errorf("remind вернул ошибку %v, want содержащую ошибку отправки %v", err, errBox)
	}
	return nil
}

type call struct{ to, subject string }

// spyMailer запоминает вызовы и падает на адресе failFor.
type spyMailer struct {
	calls   []call
	failFor string
	err     error
}

func (s *spyMailer) Send(to, subject string) error {
	s.calls = append(s.calls, call{to, subject})
	if to == s.failFor {
		return s.err
	}
	return nil
}
