package l04httpclient

import (
	"net/http"
	"time"
)

// NewClient возвращает клиент для походов в один внешний API.
//
//   - Общий таймаут запроса Client.Timeout = timeout (включая чтение тела).
//     У http.DefaultClient таймаута нет вовсе: зависший сервер повесит горутину навсегда.
//   - Транспорт: КОПИЯ http.DefaultTransport (метод Clone), а не он сам.
//     Менять поля DefaultTransport нельзя: он общий для всей программы
//     и всех библиотек в ней. Копия сохраняет прокси из окружения,
//     таймауты дозвона и TLS, HTTP/2.
//   - В копии: MaxIdleConnsPerHost = 32 (по умолчанию 2: под нагрузкой
//     соединения к одному хосту постоянно закрываются и открываются заново)
//     и ResponseHeaderTimeout = timeout (сколько ждать заголовков ответа).
func NewClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConnsPerHost = 32
	tr.ResponseHeaderTimeout = timeout
	return &http.Client{Timeout: timeout, Transport: tr}
}
