package l02tools

// Greet возвращает приветствие для name.
// Пустое имя: "Привет, незнакомец!".
func Greet(name string) string {
	if name == "" {
		name = "незнакомец"
	}
	return "Привет, " + name + "!"
}
