package l08structs

import "time"

// Timestamps: время создания и последнего изменения записи.
// Встраивается в сущности, у которых эти поля должны быть.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Touch отмечает изменение: UpdatedAt = now.
func (t *Timestamps) Touch(now time.Time) {
	panic("TODO: реализуй Touch")
}

// Article: статья блога. Timestamps встроена, поэтому у Article есть
// поля CreatedAt, UpdatedAt и метод Touch, а в JSON эти поля лежат
// на верхнем уровне рядом с id и title.
type Article struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Timestamps
}

// NewArticle создаёт статью: CreatedAt и UpdatedAt равны now.
func NewArticle(id int, title string, now time.Time) Article {
	panic("TODO: реализуй NewArticle")
}

// Rename меняет заголовок и отмечает изменение через встроенный Touch.
// Пустой заголовок не принимается: статья не меняется, возвращается false.
func (a *Article) Rename(title string, now time.Time) bool {
	panic("TODO: реализуй Rename")
}
