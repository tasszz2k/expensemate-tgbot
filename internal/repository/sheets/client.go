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

// UpdateUserEntered writes values with USER_ENTERED input option so formulas are interpreted
func (c *Client) UpdateUserEntered(ctx context.Context, spreadsheetID, writeRange string, values [][]interface{}) error {
	vr := &sheets.ValueRange{Values: values}
	_, err := c.svc.Spreadsheets.Values.Update(spreadsheetID, writeRange, vr).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("updating range %s: %w", writeRange, err)
	}
	return nil
}

// BatchUpdate executes structural spreadsheet changes (e.g., duplicate sheet)
func (c *Client) BatchUpdate(ctx context.Context, spreadsheetID string, requests []*sheets.Request) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	req := &sheets.BatchUpdateSpreadsheetRequest{Requests: requests}
	resp, err := c.svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("batch update: %w", err)
	}
	return resp, nil
}

// ClearValues clears cell values in a range while preserving formatting
func (c *Client) ClearValues(ctx context.Context, spreadsheetID, clearRange string) error {
	_, err := c.svc.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("clearing range %s: %w", clearRange, err)
	}
	return nil
}

// GetUnformatted reads raw unformatted values from a range (numbers without currency symbols, etc.)
func (c *Client) GetUnformatted(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	resp, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, readRange).
		ValueRenderOption("UNFORMATTED_VALUE").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("reading unformatted values from %s: %w", readRange, err)
	}
	return resp, nil
}

// GetFormulas reads raw formulas from a range (returns formula strings instead of computed values)
func (c *Client) GetFormulas(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	resp, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, readRange).
		ValueRenderOption("FORMULA").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("reading formulas from %s: %w", readRange, err)
	}
	return resp, nil
}

// GetSheetID returns the numeric sheet ID for a given sheet name
func (c *Client) GetSheetID(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
	allSheets, err := c.GetSheets(ctx, spreadsheetID)
	if err != nil {
		return 0, err
	}
	for _, s := range allSheets {
		if s.Properties.Title == sheetName {
			return s.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("sheet %q not found", sheetName)
}
