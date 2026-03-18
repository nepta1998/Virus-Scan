// Package service
package service

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type FileScanService struct{}

func NewFileScanService() (*FileScanService, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}
	apiKey := os.Getenv("FILESCAN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("VIRUSTOTAL_API_KEY is empty")
	}

	return &FileScanService{}, nil
}

// func (s *FileScanService) ScanFile(file *os.File, progress chan <- float32, params map[string]string) (string, error) {
//     scanner := s.client.NewFileScanner()
//     scan, err := scanner.ScanFileWithParameters(file, progress, params)
//     close(progress)
//     if err != nil {
//       return "", err
//     }
//     return scan.ID(), nil
// }
//
// func (s *FileScanService) ScanURL(url string) (string, error) {
// 	scanner := s.client.NewURLScanner()
// 	scan, err := scanner.Scan(url)
// 	if err != nil {
// 		return "", err
// 	}
// 	return scan.ID(), nil
// }
