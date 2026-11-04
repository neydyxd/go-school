package l05packages

// SplitPathVersion разбирает путь модуля по правилу совместимости импорта.
//
// Если последний элемент пути имеет вид /vN, где N целое число >= 2 без
// ведущих нулей, верни путь без этого элемента, N и true:
//
//	"github.com/jackc/pgx/v5" -> "github.com/jackc/pgx", 5, true
//
// Если суффикса версии нет, модуль имеет мажорную версию 0 или 1:
// верни весь путь, 1 и true ("example.com/lib" -> "example.com/lib", 1, true).
// Элемент, который только похож на версию ("v2x", "v", "version"), суффиксом не считается.
//
// Некорректно (верни "", 0, false): пустой путь; суффиксы /v0 и /v1
// (для них суффикс не пишут); ведущие нули (/v02); суффикс без пути перед ним ("/v2").
func SplitPathVersion(path string) (prefix string, major int, ok bool) {
	panic("TODO: реализуй SplitPathVersion")
}

// CompatibleVersion сообщает, может ли версия version (например "v2.3.1")
// принадлежать модулю с путём path:
//   - путь без суффикса: допустимы только версии v0.x.y и v1.x.y;
//   - путь с суффиксом /vN: только версии vN.x.y.
//
// Версия должна начинаться с "v" и мажорного числа, за которым идёт ".";
// иначе, как и при некорректном пути, верни false.
func CompatibleVersion(path, version string) bool {
	panic("TODO: реализуй CompatibleVersion")
}
