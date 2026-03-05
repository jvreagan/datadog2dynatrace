package dynatrace

import "fmt"

// CreateNotebook creates a notebook in Dynatrace.
func (c *Client) CreateNotebook(nb *DynatraceNotebook) error {
	_, err := c.post("/api/v2/notebooks", nb)
	if err != nil {
		return fmt.Errorf("creating notebook: %w", err)
	}
	return nil
}
