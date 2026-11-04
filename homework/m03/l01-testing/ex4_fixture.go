package l01testing

import "testing"

// WriteFixture это хелпер для тестов: создаёт файл name с содержимым content
// во временном каталоге теста и возвращает полный путь к файлу.
//
// Требования:
//   - вызови t.Helper();
//   - каталог бери из t.TempDir(): тогда он удалится сам, когда тест закончится
//     (можно и через os.MkdirTemp + t.Cleanup, но зачем, если есть готовое);
//   - если файл не записался (например, имя содержит нулевой байт), останавливай
//     тест через t.Fatalf: дальше без фикстуры идти бессмысленно.
func WriteFixture(t testing.TB, name, content string) string {
	panic("TODO: реализуй WriteFixture")
}
