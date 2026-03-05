package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ResourceGroup represents a group of selectable resources.
type ResourceGroup struct {
	Type  string
	Label string
	Items []ResourceItem
}

// ResourceItem represents a single selectable resource.
type ResourceItem struct {
	ID   string
	Name string
}

// SelectResources shows an interactive checkbox selector and returns selected resource IDs grouped by type.
func SelectResources(groups []ResourceGroup) (map[string][]string, error) {
	if len(groups) == 0 {
		return map[string][]string{}, nil
	}

	m := newSelectorModel(groups)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running selector: %w", err)
	}

	result := finalModel.(selectorModel)
	if result.cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}

	return result.getSelected(), nil
}

type selectableItem struct {
	groupIdx int
	itemIdx  int
	selected bool
	isHeader bool
	label    string
	id       string
	resType  string
}

type selectorModel struct {
	items     []selectableItem
	cursor    int
	cancelled bool
	done      bool
}

func newSelectorModel(groups []ResourceGroup) selectorModel {
	var items []selectableItem

	for gi, g := range groups {
		// Add group header
		items = append(items, selectableItem{
			groupIdx: gi,
			isHeader: true,
			label:    fmt.Sprintf("%s (%d)", g.Label, len(g.Items)),
			resType:  g.Type,
		})

		// Add items
		for ii, item := range g.Items {
			name := item.Name
			if name == "" {
				name = item.ID
			}
			items = append(items, selectableItem{
				groupIdx: gi,
				itemIdx:  ii,
				selected: true, // Select all by default
				label:    name,
				id:       item.ID,
				resType:  g.Type,
			})
		}
	}

	return selectorModel{items: items}
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
			// Skip headers
			if m.items[m.cursor].isHeader {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.items) - 1
				}
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
			// Skip headers
			if m.items[m.cursor].isHeader {
				m.cursor++
				if m.cursor >= len(m.items) {
					m.cursor = 0
				}
			}
		case " ":
			if !m.items[m.cursor].isHeader {
				m.items[m.cursor].selected = !m.items[m.cursor].selected
			}
		case "a":
			// Toggle all in current group
			if m.cursor < len(m.items) {
				groupIdx := m.items[m.cursor].groupIdx
				allSelected := true
				for _, item := range m.items {
					if item.groupIdx == groupIdx && !item.isHeader && !item.selected {
						allSelected = false
						break
					}
				}
				for i := range m.items {
					if m.items[i].groupIdx == groupIdx && !m.items[i].isHeader {
						m.items[i].selected = !allSelected
					}
				}
			}
		case "A":
			// Toggle all
			allSelected := true
			for _, item := range m.items {
				if !item.isHeader && !item.selected {
					allSelected = false
					break
				}
			}
			for i := range m.items {
				if !m.items[i].isHeader {
					m.items[i].selected = !allSelected
				}
			}
		}
	}

	return m, nil
}

func (m selectorModel) View() string {
	if m.done {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Select resources to migrate:\n")
	sb.WriteString("(space=toggle, a=toggle group, A=toggle all, enter=confirm, q=cancel)\n\n")

	for i, item := range m.items {
		if item.isHeader {
			sb.WriteString(fmt.Sprintf("\n  %s\n", item.label))
			continue
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		check := "[ ]"
		if item.selected {
			check = "[x]"
		}

		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, item.label))
	}

	return sb.String()
}

func (m selectorModel) getSelected() map[string][]string {
	result := make(map[string][]string)
	for _, item := range m.items {
		if !item.isHeader && item.selected {
			result[item.resType] = append(result[item.resType], item.id)
		}
	}
	return result
}
