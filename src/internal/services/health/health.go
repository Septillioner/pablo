package health

import (
	"fmt"
	"net/http"
	"time"
)

func Check(url string, timeout time.Duration) error {
	if url == "" {
		return nil
	}

	fmt.Printf("Performing health check on %s (timeout: %s)\n", url, timeout)

	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Println("Health check passed!")
			return nil
		}

		if err != nil {
			fmt.Printf("Health check failed (retrying): %v\n", err)
		} else {
			fmt.Printf("Health check returned status %d (retrying)\n", resp.StatusCode)
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("health check timed out after %s", timeout)
}
