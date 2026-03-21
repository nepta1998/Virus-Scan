// Package tui
package tui

import (
	"fmt"
	"os"
	"time"

	"virusscan/internal/models"
	"virusscan/internal/service"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type setProgramMsg *tea.Program

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}
type tableDataMsg []table.Row

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type mode int

const (
	modeList mode = iota
	modePicker
	modeInput
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
	vtservice    *service.VirusTotalService
	spinner      spinner.Model
	loading      bool
	input        textinput.Model
}

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
		m.table.SetHeight(msg.Height - v - 10)
		m.table.SetWidth(msg.Width - h)
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
				it, ok := m.list.SelectedItem().(item)
				if ok {
					if it.title == "Scan File" {
						m.mode = modePicker
						// Importante: al entrar al picker, pedimos que se inicialice/refresque
						return m, m.filepicker.Init()
					}
					if it.title == "Scan Url" {
						m.mode = modeInput
						m.input.Focus()
						return m, textinput.Blink
					}
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
			return m, ScanFileCmd(file, m.p, m.vtservice)
		}
	case modeInput:
		var cmdInput tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.mode = modeList
				return m, nil
			}
		}

		m.input, cmdInput = m.input.Update(msg)
		return m, cmdInput

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
			m.loading = true
			return m, tea.Batch(m.loadAnalysisCmd(), m.spinner.Tick)
		}
	case modeResult:
		switch msg := msg.(type) {
		case tableDataMsg:
			m.table.SetRows(msg)
			m.table.Focus()
			m.loading = false
			return m, nil
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case tea.KeyPressMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "q":
				m.mode = modePicker
				return m, nil
			case "enter":
				return m, nil
			}
		}
		var tableCmd tea.Cmd
		m.table, tableCmd = m.table.Update(msg)
		return m, tableCmd
	}

	return m, tea.Batch(cmds...)
}

func (m model) loadAnalysisCmd() tea.Cmd {
	return func() tea.Msg {
		// 1. Pedir el análisis
		analysis, err := m.vtservice.GetAnalysis(m.analysisID)
		if err != nil {
			return tableDataMsg([]table.Row{{"ERROR", err.Error()}})
		}

		// 2. Verificar el estado global del análisis
		status, _ := analysis.Get("status")
		if status == "queued" || status == "in-progress" {
			// Si no ha terminado, esperamos 3 segundos y reintentamos
			time.Sleep(3 * time.Second)
			return m.loadAnalysisCmd()() // Reintento síncrono dentro del Cmd
		}

		// 3. Procesar resultados si ya está "completed"
		var rows []table.Row
		results, _ := analysis.Get("results")
		if r, ok := results.(map[string]interface{}); ok {
			for engine, data := range r {
				resText := "limpio" // Valor por defecto
				if d, ok := data.(map[string]interface{}); ok {
					// 1. Intentamos obtener 'result'
					if v, ok := d["result"]; ok && v != nil && v != "" {
						resText = fmt.Sprintf("%v", v)
					} else {
						if cat, ok := d["category"]; ok && cat != nil {
							resText = fmt.Sprintf("%v", cat)
						}
					}
				}
				rows = append(rows, table.Row{engine, resText})
			}
		}

		if len(rows) == 0 {
			time.Sleep(2 * time.Second)
			// return m.loadAnalysisCmd()()
		}
		return tableDataMsg(rows)
	}
}

func (m model) View() tea.View {
	var content string
	switch m.mode {
	case modeList:
		content = m.list.View()
	case modePicker:
		content = m.filepicker.View()
	case modeScanning:
		content = fmt.Sprintf(
			"\n  %s\n\n  %s\n\n  %s",
			lipgloss.NewStyle().Bold(true).Render(m.status),
			m.progress.View(),
			lipgloss.NewStyle().Italic(true).Render("Presiona q para cancelar"),
		)
	case modeResult:
		if m.loading {
			content = fmt.Sprintf("\n  %s Cargando resultados de VirusTotal...", m.spinner.View())
		} else {
			content = fmt.Sprintf(
				"\n  %s\n\n  %s\n",
				m.table.View(),
				m.table.HelpView(),
			)
		}
	case modeInput:
		content = fmt.Sprintf(
			"\n  %s\n\n  %s\n\n  %s",
			lipgloss.NewStyle().Bold(true).Render("Introduce la URL para escanear:"),
			m.input.View(),
			lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")).Render("(Presiona Enter para escanear o Esc para volver)"),
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
		{Title: "Engine", Width: 30},
		{Title: "Result", Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
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
	vtservice, _ := service.NewVirusTotalService()
	sp := spinner.New()
	sp.Spinner = spinner.Dot                                         // Define el tipo de puntos
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Color rosa

	ti := textinput.New()
	ti.Placeholder = "http://example.com"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return model{
		mode:       modeList,
		list:       list,
		filepicker: fp,
		progress:   pg,
		table:      t,
		vtservice:  vtservice,
		spinner:    sp,
		input:      ti,
	}
}

func (m model) NewView() {
	// f, err := tea.LogToFile("debug.log", "debug")
	// if err == nil {
	// 	defer f.Close()
	// }
	programInstance := tea.NewProgram(m)
	go func() {
		programInstance.Send(setProgramMsg(programInstance))
	}()
	if _, err := programInstance.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
