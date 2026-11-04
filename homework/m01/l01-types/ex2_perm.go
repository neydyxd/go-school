package l01types

// Perm это набор прав доступа в виде битовых флагов.
type Perm uint8

// TODO: объяви Read, Write и Exec через iota так, чтобы каждое право
// занимало свой бит: Read = 1, Write = 2, Exec = 4.
// Сейчас все три равны нулю, и тесты это заметят.
const (
	Read  Perm = 0
	Write Perm = 0
	Exec  Perm = 0
)

// Has сообщает, установлены ли в p ВСЕ биты из flag.
// Has(Read|Write, Read) == true, Has(Read, Read|Write) == false.
// Пустой flag (0) считается подмножеством любого набора: true.
func Has(p, flag Perm) bool {
	panic("TODO: реализуй Has")
}

// Format возвращает права в стиле ls: три символа "rwx",
// на месте отсутствующего права стоит '-'.
// Format(Read|Exec) == "r-x", Format(0) == "---".
// Биты старше Exec игнорируются.
func Format(p Perm) string {
	panic("TODO: реализуй Format")
}
