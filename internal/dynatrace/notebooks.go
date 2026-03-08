package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateNotebook creates a notebook in Dynatrace.
func (c *Client) CreateNotebook(nb *DynatraceNotebook) error {
	_, err := c.post("/api/v2/notebooks", nb)
	if err != nil {
		return fmt.Errorf("creating notebook: %w", err)
	}
	return nil
}

// ListNotebookNames returns the names of all existing notebooks.
func (c *Client) ListNotebookNames() ([]string, error) {
	data, err := c.get("/api/v2/notebooks")
	if err != nil {
		return nil, fmt.Errorf("listing notebooks: %w", err)
	}
	var resp struct {
		Notebooks []struct {
			Name string `json:"name"`
		} `json:"notebooks"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing notebook list: %w", err)
	}
	names := make([]string, len(resp.Notebooks))
	for i, nb := range resp.Notebooks {
		names[i] = nb.Name
	}
	return names, nil
}
