package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kerimcanbalkan/mp3-player/internal/audio"
)

type model struct {
	files    []string
	cursor   int
	selected string
	ready    bool
	viewport viewport.Model
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	songStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Width(200).Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).Padding(0, 1)
	}()

	choosenSongStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Width(200).Bold(true).Background(lipgloss.AdaptiveColor{Light: "14", Dark: "9"}).Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

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

func (m model) renderSongList() string {
	var content string
	for i, file := range m.files {
		musicFile := filepath.Base(file)
		ext := filepath.Ext(musicFile)
		song := strings.TrimSuffix(musicFile, ext)

		if m.cursor == i {
			song = choosenSongStyle.Render(song)
		} else {
			song = songStyle.Render(song)
		}

		content += fmt.Sprintf("%s\n", song)
	}
	return content
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.renderSongList())

			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

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
				m.viewport.SetContent(m.renderSongList())
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				m.selected = m.files[m.cursor]
				m.viewport.SetContent(m.renderSongList())
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
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render("Select Song To Play!")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprint("Press q to exit"))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
