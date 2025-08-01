package gsheetclients

import (
	"context"
	"log"

	"expensemate-tgbot/pkg/configs"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var gSheets GSheeter

type GSheeter interface {
	Get(ctx context.Context, spreadsheetId string, readRange string) (*sheets.ValueRange, error)
	Update(
		ctx context.Context,
		spreadsheetId string,
		writeRange string,
		vr *sheets.ValueRange,
	) (*sheets.UpdateValuesResponse, error)
	GetValue(ctx context.Context, spreadsheetId string, readRange string) (string, error)
	GetListOfSheets(ctx context.Context, spreadsheetId string) ([]*sheets.Sheet, error)
}

type GSheets struct {
	Svc *sheets.Service
}

func newGSheets(svc *sheets.Service) GSheeter {
	return &GSheets{
		Svc: svc,
	}
}

func init() {
	ctx := context.Background()
	credential := configs.Get().GoogleApis.Credentials

	jwtConfig := &jwt.Config{
		Email:      credential.ClientEmail,
		PrivateKey: []byte(credential.PrivateKey),
		Scopes:     []string{sheets.SpreadsheetsScope},
		TokenURL:   credential.TokenURI,
	}

	// Obtain an OAuth2 token for the service account
	tokenSource := jwtConfig.TokenSource(ctx)
	_, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Unable to obtain token: %v", err)
	}

	// Create a new Sheets service with the token
	svc, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}

	gSheets = newGSheets(svc)
	return
}

func GetInstance() GSheeter {
	return gSheets
}

func (g *GSheets) Get(
	ctx context.Context,
	spreadsheetId string,
	readRange string,
) (*sheets.ValueRange, error) {
	return g.Svc.Spreadsheets.Values.Get(spreadsheetId, readRange).Context(ctx).Do()
}

func (g *GSheets) Update(
	ctx context.Context,
	spreadsheetId string,
	writeRange string,
	valueRange *sheets.ValueRange,
) (*sheets.UpdateValuesResponse, error) {
	return g.Svc.Spreadsheets.Values.Update(
		spreadsheetId,
		writeRange,
		valueRange,
	).ValueInputOption("RAW").Context(ctx).Do()
}

func (g *GSheets) GetValue(ctx context.Context, spreadsheetId string, readRange string) (
	string,
	error,
) {
	resp, err := g.Get(ctx, spreadsheetId, readRange)
	if err != nil {
		return "", err
	}
	if len(resp.Values) == 0 {
		return "", nil
	}
	return resp.Values[0][0].(string), nil
}

func (g *GSheets) GetListOfSheets(ctx context.Context, spreadsheetId string) (
	[]*sheets.Sheet,
	error,
) {
	resp, err := g.Svc.Spreadsheets.Get(spreadsheetId).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return resp.Sheets, nil
}
