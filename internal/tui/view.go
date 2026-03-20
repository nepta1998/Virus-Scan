// Package tui
package tui

import (
	"fmt"
	"os"

	"virusscan/internal/models"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type setProgramMsg *tea.Program

var (
	docStyle        = lipgloss.NewStyle().Margin(1, 2)
	programInstance *tea.Program
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type mode int

const (
	modeList mode = iota
	modePicker
	modeScanning
	modeResult
)

type model struct {
	mode         mode
	list         list.Model
	filepicker   filepicker.Model
	selectedFile string
	progress     progress.Model
	status       string
	p            *tea.Program
	table        table.Model
	analysisID   string
}

// type clearErrorMsg struct{}
//
// func clearErrorAfter(t time.Duration) tea.Cmd {
// 	return tea.Tick(t, func(_ time.Time) tea.Msg {
// 		return clearErrorMsg{}
// 	})
// }

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // Usaremos una lista de comandos

	// --- LÓGICA GLOBAL ---
	switch msg := msg.(type) {
	case setProgramMsg:
		m.p = msg
		return m, nil
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		// Actualizamos ambos SIEMPRE
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.filepicker.SetHeight(msg.Height - v - 4)
		m.progress.SetWidth(msg.Width - h - 4)
		return m, nil
	}

	// --- LÓGICA POR MODO ---
	switch m.mode {
	case modeList:
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)

		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				if it, ok := m.list.SelectedItem().(item); ok && it.title == "Scan File" {
					m.mode = modePicker
					// Importante: al entrar al picker, pedimos que se inicialice/refresque
					return m, m.filepicker.Init()
				}
			}
		}

	case modePicker:
		var pickerCmd tea.Cmd
		m.filepicker, pickerCmd = m.filepicker.Update(msg)
		cmds = append(cmds, pickerCmd)

		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.mode = modeList
				return m, nil
			case "backspace":
				return m, pickerCmd
			}
		}

		if ok, path := m.filepicker.DidSelectFile(msg); ok {
			m.selectedFile = path
			file, _ := os.Open(path)
			m.mode = modeScanning
			m.status = "Subiendo archivo..."
			// m.mode = modeList
			return m, ScanFileCmd(file, m.p)
		}
	case modeScanning:
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "q":
				m.mode = modePicker
				return m, nil
			}
		}
		switch msg := msg.(type) {
		case models.VTProgress:
			// 1. Calculamos el porcentaje (debe ser entre 0.0 y 1.0)
			pct := float64(msg)
			m.status = fmt.Sprintf("Escaneando... %.0f%%", pct*100)
			var cmdProgress tea.Cmd
			cmdProgress = m.progress.SetPercent(pct)
			return m, cmdProgress
		case progress.FrameMsg:
			var cmdFrame tea.Cmd
			m.progress, cmdFrame = m.progress.Update(msg)
			return m, cmdFrame
		case models.VTResult:
			if msg.Err != nil {
				m.status = "Error: " + msg.Err.Error()
				return m, nil
			}
			m.analysisID = msg.ID
			m.status = fmt.Sprintf("¡ÉXITO! ID Recibido: %s", msg.ID)
			m.mode = modeResult
			return m, nil
		}
	case modeResult:
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "q":
				m.mode = modePicker
				return m, nil
			case "enter":
				return m, tea.Batch(
					tea.Printf("Let's go to %s! %s", m.table.SelectedRow()[1]),
				)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var content string
	switch m.mode {
	case modeList:
		content = m.list.View()
	case modePicker:
		content = m.filepicker.View()
	case modeScanning:
		// Estilizamos el progreso y el mensaje
		content = fmt.Sprintf(
			"\n  %s\n\n  %s\n\n  %s",
			lipgloss.NewStyle().Bold(true).Render(m.status),
			m.progress.View(),
			lipgloss.NewStyle().Italic(true).Render("Presiona q para cancelar"),
		)
	case modeResult:
		content = fmt.Sprintf(
			"\n  %s\n\n  %s\n",
			m.table.View(),
			m.table.HelpView(),
		)
	}

	v := tea.NewView(docStyle.Render(content))
	v.AltScreen = true
	return v
}

func NewModel() model {
	items := []list.Item{
		item{title: "Scan File", desc: "Scan File"},
		item{title: "Scan Url", desc: "Scan Url"},
	}
	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Opciones"
	fp := filepicker.New()
	userHome, _ := os.UserHomeDir()
	fp.CurrentDirectory = userHome
	pg := progress.New(progress.WithDefaultBlend())
	columns := []table.Column{
		{Title: "Rank", Width: 4},
		{Title: "City", Width: 10},
		{Title: "Country", Width: 10},
		{Title: "Population", Width: 10},
	}

	rows := []table.Row{
		{"1", "Tokyo", "Japan", "37,274,000"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithWidth(42),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return model{
		mode:       modeList,
		list:       list,
		filepicker: fp,
		progress:   pg,
		table:      t,
	}
}

func (m model) NewView() {
	programInstance := tea.NewProgram(m)
	go func() {
		programInstance.Send(setProgramMsg(programInstance))
	}()
	if _, err := programInstance.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
