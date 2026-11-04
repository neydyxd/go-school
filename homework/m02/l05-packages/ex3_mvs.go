package l05packages

// Module: конкретная версия модуля, как в строке require.
type Module struct {
	Path    string
	Version string
}

// BuildList реализует Minimal Version Selection в простейшем виде
// (алгоритм 1 из статьи Расса Кокса «Minimal Version Selection»):
//
//  1. Обойди граф требований, начиная с root: graph[m] это список
//     require из go.mod модуля m. Посети каждую достижимую пару
//     путь@версия ровно один раз (в графе бывают циклы).
//  2. Для каждого пути модуля выбери максимальную из встреченных версий
//     (сравнивай через CompareVersions).
//
// Результат: мапа путь -> выбранная версия, root тоже входит.
// Модуль без записи в graph просто не имеет зависимостей.
// Если встретилась некорректная версия, верни nil и ошибку с ErrInvalidVersion.
func BuildList(root Module, graph map[Module][]Module) (map[string]string, error) {
	panic("TODO: реализуй BuildList")
}
