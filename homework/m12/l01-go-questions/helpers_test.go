package l01goquestions

import (
	"testing"
	"time"
)

// watchdog роняет тест с понятным сообщением, если он висит дольше 5 секунд
// реального времени (synctest не замечает горутин, ждущих на мьютексе).
func watchdog(t *testing.T, hint string) (stop func() bool) {
	name := t.Name()
	return time.AfterFunc(5*time.Second, func() {
		panic("тест " + name + " завис: " + hint)
	}).Stop
}
