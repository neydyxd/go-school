package l04brokers

// Deduper отсеивает повторные доставки по id сообщения (message_id в
// RabbitMQ, ключ идемпотентности в Kafka). Помнит последние capacity РАЗНЫХ
// id: когда места нет, забывает самый давний. Повтор старше окна пройдёт как
// новый, поэтому окно подбирают больше, чем глубина возможных повторов, а
// окончательную защиту держит уникальный ключ в базе.
//
// Нужно O(1) на вызов: карта для поиска и кольцевой буфер (или очередь) для
// порядка вытеснения. Не потокобезопасен: один Deduper на одного потребителя.
type Deduper struct {
	capacity int
	seen     map[string]struct{}
	order    []string // кольцевой буфер id в порядке появления
	head     int      // индекс самого давнего id, когда буфер полон
}

// NewDeduper создаёт окно на capacity id. capacity < 1 считай за 1.
func NewDeduper(capacity int) *Deduper {
	capacity = max(capacity, 1)
	return &Deduper{capacity: capacity, seen: make(map[string]struct{}, capacity)}
}

// Seen возвращает true, если id уже был в окне (это дубль, сообщение надо
// подтвердить и пропустить). Иначе запоминает id и возвращает false.
// Повторный Seen не продлевает жизнь id: вытесняется по времени первого появления.
func (d *Deduper) Seen(id string) bool {
	if _, ok := d.seen[id]; ok {
		return true
	}
	if len(d.order) < d.capacity {
		d.order = append(d.order, id)
	} else {
		delete(d.seen, d.order[d.head])
		d.order[d.head] = id
		d.head = (d.head + 1) % d.capacity
	}
	d.seen[id] = struct{}{}
	return false
}
