package service

import (
	"fmt"
	"os"

	vt "github.com/VirusTotal/vt-go"
	"github.com/joho/godotenv"
)

type VirusTotalService struct {
	client *vt.Client
}

func NewVirusTotalService() (*VirusTotalService, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}
	apiKey := os.Getenv("VIRUSTOTAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("VIRUSTOTAL_API_KEY is empty")
	}
	client := vt.NewClient(apiKey)

	return &VirusTotalService{client}, nil
}

func (s *VirusTotalService) ScanFile(file *os.File, progress chan <- float32, params map[string]string) (string, error) {
    scanner := s.client.NewFileScanner()
    scan, err := scanner.ScanFileWithParameters(file, progress, params)
    close(progress)
    if err != nil {
      return "", err
    }
    return scan.ID(), nil
}

func (s *VirusTotalService) ScanURL(url string) (string, error) {
	scanner := s.client.NewURLScanner()
	scan, err := scanner.Scan(url)
	if err != nil {
		return "", err
	}
	return scan.ID(), nil
}

func (s *VirusTotalService) GetAnalysis(analysisID string) (*vt.Object, error) {
	analysis, err := s.client.GetObject(vt.URL("analyses/%s", analysisID))
	if err != nil {
		return nil, err
	}
	return analysis, nil
}
