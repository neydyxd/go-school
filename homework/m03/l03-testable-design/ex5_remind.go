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
// Не вызывай Remind внутри CheckRemind и не сравнивай remind с ней: проверка
// должна опираться на правила выше, а не на эталон, иначе это не тест.
//
// Тест прогонит CheckRemind против правильных реализаций (должно быть nil)
// и против мутантов из ex5_remind_test.go (каждого надо поймать).
func CheckRemind(remind RemindFunc) error {
	panic("TODO: реализуй CheckRemind")
}
