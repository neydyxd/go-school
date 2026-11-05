package l04httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientConfig(t *testing.T) {
	def := http.DefaultTransport.(*http.Transport)
	beforeIdle, beforeRHT := def.MaxIdleConnsPerHost, def.ResponseHeaderTimeout
	c := NewClient(3 * time.Second)
	if c == nil || c == http.DefaultClient {
		t.Fatal("NewClient вернул nil или http.DefaultClient, want новый клиент")
	}
	if c.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want 3s", c.Timeout)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", c.Transport)
	}
	if tr == http.DefaultTransport {
		t.Fatal("Transport это сам http.DefaultTransport, want его копию (Clone)")
	}
	if tr.MaxIdleConnsPerHost != 32 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 32", tr.MaxIdleConnsPerHost)
	}
	if tr.ResponseHeaderTimeout != 3*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 3s", tr.ResponseHeaderTimeout)
	}
	if tr.Proxy == nil || tr.DialContext == nil || tr.TLSHandshakeTimeout == 0 || tr.IdleConnTimeout == 0 {
		t.Error("у транспорта нет Proxy/DialContext/TLSHandshakeTimeout/IdleConnTimeout: собран с нуля вместо Clone")
	}
	if def.MaxIdleConnsPerHost != beforeIdle || def.ResponseHeaderTimeout != beforeRHT {
		t.Error("http.DefaultTransport изменён: его нельзя трогать, он общий для всей программы")
	}
}

func TestNewClientTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer srv.Close()
	defer close(release)

	c := NewClient(100 * time.Millisecond)
	done := make(chan error, 1)
	go func() {
		resp, err := c.Get(srv.URL)
		if err == nil {
			resp.Body.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("запрос к зависшему серверу завершился без ошибки, want таймаут")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("запрос к зависшему серверу висит больше 3с при таймауте 100мс")
	}
}
