package http

import (
	"net/url"
	"strings"
)

// IsValidURL checks if a string is a valid URL
func IsValidURL(u string) bool {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// IsValidGoogleSheetsURL validates a Google Sheets URL
func IsValidGoogleSheetsURL(u string) bool {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return false
	}

	if parsedURL.Host != "docs.google.com" {
		return false
	}

	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 4 || pathParts[1] != "spreadsheets" || pathParts[2] != "d" {
		return false
	}

	docID := pathParts[3]
	return docID != ""
}

// GetGoogleSheetsDocID extracts the document ID from a Google Sheets URL
func GetGoogleSheetsDocID(u string) string {
	if !IsValidGoogleSheetsURL(u) {
		return ""
	}
	parsedURL, _ := url.Parse(u)
	pathParts := strings.Split(parsedURL.Path, "/")
	return pathParts[3]
}
