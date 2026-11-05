package l01ginintro

import "github.com/gin-gonic/gin"

// NewRouter собирает роутер на gin.New() (без Logger и Recovery) с четырьмя маршрутами.
// Все ответы в JSON. На ошибку отвечай одним объектом {"error": "<текст>"} и сразу выходи
// из обработчика: второй c.JSON после первого допишет второй объект в то же тело.
//
//   - GET /ping: 200 {"message": "pong"}.
//   - GET /users/:id: id это целое число больше нуля (strconv.ParseInt),
//     тогда 200 {"id": 42} (число, не строка). Иначе 400.
//   - GET /files/*path: 200 {"path": "docs/a.txt"} для /files/docs/a.txt.
//     Ведущий слеш из параметра убери. Пустой путь (/files/) это 400.
//   - GET /search?q=go&limit=5: q обязателен (пустой или из одних пробелов это 400),
//     пробелы по краям обрежь. limit по умолчанию 10, допустимо 1..50, не число или
//     вне диапазона это 400. Ответ 200 {"q": "go", "limit": 5}.
func NewRouter() *gin.Engine {
	panic("TODO: реализуй NewRouter")
}
