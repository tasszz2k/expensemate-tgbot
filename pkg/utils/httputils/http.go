package httputils

import (
	"net/url"
	"strings"
)

func IsValidURL(u string) bool {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func IsValidGoogleSheetsURL(u string) bool {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return false
	}

	// Check if the host is "docs.google.com"
	if parsedURL.Host != "docs.google.com" {
		return false
	}

	// Check if the path contains "/spreadsheets/d/"
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 4 || pathParts[1] != "spreadsheets" || pathParts[2] != "d" {
		return false
	}

	// Check if the path has a valid document ID
	docID := pathParts[3]
	if docID == "" {
		return false
	}

	return true
}

func GetGoogleSheetsDocID(u string) string {
	if !IsValidGoogleSheetsURL(u) {
		return ""
	}
	parsedURL, _ := url.Parse(u)
	pathParts := strings.Split(parsedURL.Path, "/")
	return pathParts[3]
}
