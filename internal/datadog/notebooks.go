package datadog

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetNotebooks retrieves all notebooks.
func (c *Client) GetNotebooks() ([]Notebook, error) {
	data, err := c.get("/api/v1/notebooks")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []struct {
			ID         int64             `json:"id"`
			Attributes NotebookAttributes `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing notebooks: %w", err)
	}

	var notebooks []Notebook
	for _, nb := range resp.Data {
		notebooks = append(notebooks, Notebook{
			ID:       nb.ID,
			Name:     nb.Attributes.Name,
			Author:   nb.Attributes.Author,
			Cells:    nb.Attributes.Cells,
			Created:  nb.Attributes.Created,
			Modified: nb.Attributes.Modified,
		})
	}
	return notebooks, nil
}

type NotebookAttributes struct {
	Name     string         `json:"name"`
	Author   NotebookAuthor `json:"author"`
	Cells    []NotebookCell `json:"cells"`
	Created  time.Time      `json:"created"`
	Modified time.Time      `json:"modified"`
}
