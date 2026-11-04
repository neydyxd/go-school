package l03testabledesign

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Profile это профиль из внешнего API.
type Profile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ErrNoProfile означает, что API ответил 404.
var ErrNoProfile = errors.New("profile not found")

// FetchProfile запрашивает профиль у внешнего API клиентом c.
func FetchProfile(c *http.Client, baseURL string, id int) (Profile, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/profiles/"+strconv.Itoa(id), nil)
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("fetch profile %d: %w", id, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return Profile{}, fmt.Errorf("fetch profile %d: %w", id, ErrNoProfile)
	default:
		return Profile{}, fmt.Errorf("fetch profile %d: unexpected status %d", id, resp.StatusCode)
	}
	var p Profile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return Profile{}, fmt.Errorf("fetch profile %d: decode: %w", id, err)
	}
	return p, nil
}
