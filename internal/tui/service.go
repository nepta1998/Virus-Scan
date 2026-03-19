package tui

import (
	"fmt"
	// "math"
	"os"

	"virusscan/internal/models"
	"virusscan/internal/service"

	tea "charm.land/bubbletea/v2"
)

func ScanFileCmd(file *os.File, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		// 1. Validamos que el programa no sea nil antes de empezar
		if p == nil {
			return models.VTResult{Err: fmt.Errorf("program instance is nil")}
		}

		progressChan := make(chan float32, 100)
		resChan := make(chan models.VTResult)

		// Goroutine: Escaneo real
		go func() {
			vtservice, _ := service.NewVirusTotalService()
			id, err := vtservice.ScanFile(file, progressChan, nil)
			resChan <- models.VTResult{ID: id, Err: err}
		}()

		// Goroutine: Escucha el progreso y lo manda a la UI vía Send()
		go func() {
			for pct := range progressChan {
				// p ya no es nil porque lo validamos arriba
				p.Send(models.VTProgress(pct/100))
			}
		}()

		// El comando principal espera y retorna el RESULTADO FINAL
		return <-resChan
	}
}


