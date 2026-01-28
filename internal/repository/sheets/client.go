package sheets

import (
	"context"
	"fmt"

	"expensemate-tgbot/internal/config"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the Google Sheets API service
type Client struct {
	svc *sheets.Service
}

// NewClient creates a new Google Sheets client
func NewClient(cfg *config.Config) (*Client, error) {
	ctx := context.Background()
	creds := cfg.GoogleAPIs.Credentials

	jwtConfig := &jwt.Config{
		Email:      creds.ClientEmail,
		PrivateKey: []byte(creds.PrivateKey),
		Scopes:     []string{sheets.SpreadsheetsScope},
		TokenURL:   creds.TokenURI,
	}

	tokenSource := jwtConfig.TokenSource(ctx)
	if _, err := tokenSource.Token(); err != nil {
		return nil, fmt.Errorf("obtaining token: %w", err)
	}

	svc, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	return &Client{svc: svc}, nil
}

// Get reads values from a spreadsheet range
func (c *Client) Get(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	resp, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading range %s: %w", readRange, err)
	}
	return resp, nil
}

// Update writes values to a spreadsheet range
func (c *Client) Update(ctx context.Context, spreadsheetID, writeRange string, values [][]interface{}) error {
	vr := &sheets.ValueRange{Values: values}
	_, err := c.svc.Spreadsheets.Values.Update(spreadsheetID, writeRange, vr).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("updating range %s: %w", writeRange, err)
	}
	return nil
}

// GetValue reads a single cell value
func (c *Client) GetValue(ctx context.Context, spreadsheetID, cell string) (string, error) {
	resp, err := c.Get(ctx, spreadsheetID, cell)
	if err != nil {
		return "", err
	}
	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return "", nil
	}

	value, ok := resp.Values[0][0].(string)
	if !ok {
		return fmt.Sprintf("%v", resp.Values[0][0]), nil
	}
	return value, nil
}

// GetSheets returns all sheets in a spreadsheet
func (c *Client) GetSheets(ctx context.Context, spreadsheetID string) ([]*sheets.Sheet, error) {
	resp, err := c.svc.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("getting spreadsheet: %w", err)
	}
	return resp.Sheets, nil
}

// Append appends rows to a spreadsheet
func (c *Client) Append(ctx context.Context, spreadsheetID, appendRange string, values [][]interface{}) error {
	vr := &sheets.ValueRange{Values: values}
	_, err := c.svc.Spreadsheets.Values.Append(spreadsheetID, appendRange, vr).
		ValueInputOption("RAW").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("appending to range %s: %w", appendRange, err)
	}
	return nil
}
