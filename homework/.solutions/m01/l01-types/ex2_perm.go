package l01types

// Perm это набор прав доступа в виде битовых флагов.
type Perm uint8

// Каждое право занимает свой бит: Read = 1, Write = 2, Exec = 4.
// Выражение 1 << iota повторяется для Write и Exec с новым значением iota.
const (
	Read Perm = 1 << iota
	Write
	Exec
)

// Has сообщает, установлены ли в p ВСЕ биты из flag.
// Has(Read|Write, Read) == true, Has(Read, Read|Write) == false.
// Пустой flag (0) считается подмножеством любого набора: true.
func Has(p, flag Perm) bool {
	return p&flag == flag
}

// Format возвращает права в стиле ls: три символа "rwx",
// на месте отсутствующего права стоит '-'.
// Format(Read|Exec) == "r-x", Format(0) == "---".
// Биты старше Exec игнорируются.
func Format(p Perm) string {
	b := []byte("---")
	if Has(p, Read) {
		b[0] = 'r'
	}
	if Has(p, Write) {
		b[1] = 'w'
	}
	if Has(p, Exec) {
		b[2] = 'x'
	}
	return string(b)
}
