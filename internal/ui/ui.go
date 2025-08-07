package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kerimcanbalkan/mp3-player/internal/audio"
)

type model struct {
	files    []string
	cursor   int
	selected string
}

func InitialModel() model {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var mp3Files []string

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			mp3Files = append(mp3Files, path)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return model{
		// Our to-do list is a grocery list
		files: mp3Files,

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: mp3Files[0],
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.selected = m.files[m.cursor]
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				m.selected = m.files[m.cursor]
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			go audio.Play(m.selected)
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Select Music to Play\n\n"

	// Iterate over our choices
	for i, file := range m.files {

		musicFile := filepath.Base(file) // "song.mp3"
		ext := filepath.Ext(musicFile)
		song := strings.TrimSuffix(musicFile, ext)

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, song)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
