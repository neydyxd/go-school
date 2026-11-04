package l01testing

// Общая механика «охоты на мутантов» для заданий, где ты пишешь тесты.
// Мутант это заведомо сломанная копия функции. Твоя таблица кейсов
// «ловит» мутанта, если хотя бы на одном кейсе он даёт неверный ответ.

// survives сообщает, что мутант прошёл все кейсы, то есть таблица его не поймала.
// check возвращает true, если на кейсе ответ мутанта совпал с ожидаемым.
// Паника мутанта считается пойманной: настоящий тест на этом кейсе упал бы.
func survives[C any](cases []C, check func(C) bool) bool {
	for _, c := range cases {
		if !safeCheck(c, check) {
			return false
		}
	}
	return true
}

func safeCheck[C any](c C, check func(C) bool) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return check(c)
}
