package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Checker struct {
	client *http.Client
}

func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Checker) Check(host string, port int, path string) (bool, string) {
	url := fmt.Sprintf("http://%s:%d%s", host, port, path)
	log.Info().Str("url", url).Msg("running health check")

	resp, err := c.client.Get(url)
	if err != nil {
		return false, fmt.Sprintf("health check request failed: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("health check returned status %d", resp.StatusCode)
	}

	return true, fmt.Sprintf("health check passed (status %d)", resp.StatusCode)
}
