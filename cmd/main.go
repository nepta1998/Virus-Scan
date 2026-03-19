package main

import (
	"virusscan/internal/tui"
)

func main() {
	tui.NewModel().NewView()
}

// import (
// 	"fmt"
// 	"os"
//
// 	"virusscan/internal/models"
// 	"virusscan/internal/service"
// )
//
// func main() {
// 	vtservice, err := service.NewVirusTotalService()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
//
// 	}
// 	file, err := os.Open("/home/neptali/projects/rust/ejemplo1/Cargo.toml")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer file.Close()
// 	progressChan := make(chan float32, 100)
// 	resChan := make(chan models.VTResult)
// 	go func() {
// 		analysisID, err := vtservice.ScanFile(file, progressChan, map[string]string{})
// 		resChan <- models.VTResult{ID: analysisID, Err: err}
// 	}()
// 	for p := range progressChan {
// 		fmt.Printf("\rProgreso: %.2f%%\n", p)
// 	}
// 	result := <-resChan
// 	if result.Err != nil {
// 		fmt.Println("\nError:", result.Err)
// 		return
// 	}
// 	analysis, err := vtservice.GetAnalysis(result.ID)
// 	if err != nil {
// 		fmt.Println("Error getting analysis:", err)
// 		return
// 	}
// 	fmt.Println(analysis)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	//
// 	// fmt.Println("Analysis ID:", analysisID)
// 	//
// 	// analysis, err := vtservice.GetAnalysis(analysisID)
// 	// if err != nil {
// 	// 	fmt.Println("Error getting analysis:", err)
// 	// 	return
// 	// }
// 	//
// 	// // fmt.Println(analysis)
// 	// fmt.Printf("%+v\n", *analysis)
// 	// // stats, _ := analysis.Get("stats")
// 	// // fmt.Println("Stats:", stats)
// 	// // results, _ := analysis.Get("results")
// 	// // fmt.Println("Results:", results)
// 	// //
// 	// status, _ := analysis.Get("Links")
// 	// fmt.Println("Status:", status)
// }
