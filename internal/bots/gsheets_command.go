package bots

import (
	"context"
	"fmt"
	"strings"

	"expensemate-tgbot/pkg/clients/gsheetclients"
	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/gsheettypes"
	"expensemate-tgbot/pkg/types/tgtypes"
	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/httputils"
	"expensemate-tgbot/pkg/utils/timeutils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (e *Expensemate) handleGSheetsCommand(
	ctx context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")

	// Check if the user configured a Google Sheets
	// if not, send a message to the user to configure it
	// if yes, send a message to the user with the configured Google Sheets

	spreadsheetMapping, err := e.getSpreadsheetMappingByUserID(ctx, incomingMessage.From.ID)
	if err != nil || spreadsheetMapping.SpreadSheetsURL == "" ||
		!httputils.IsValidURL(spreadsheetMapping.SpreadSheetsURL) {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}

	// user has configured a Google Sheets
	// show the configured Google Sheets
	msg.Text = fmt.Sprintf(
		"You have configured a Google Sheets. Click the button below to view it: %s",
		spreadsheetMapping.SpreadSheetsURL,
	)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
			tgbotapi.NewInlineKeyboardButtonURL(
				"View",
				spreadsheetMapping.SpreadSheetsURL,
			),
			tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Update Current Page",
				"gsheets:update_current_page",
			),
		),
	)

	return msg, nil
}

func (e *Expensemate) handleGSheetsCallback(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
) (tgbotapi.MessageConfig, error) {
	callbackQueryData := query.Data
	_, subCommands := tgtypes.ParseCallbackData(callbackQueryData)
	action := gsheettypes.Action(subCommands[0])
	var msg tgbotapi.MessageConfig

	switch action {
	case gsheettypes.ActionConfigure:
		msg, _ = e.handleGSheetsConfigureCallback(ctx, query)
	case gsheettypes.ActionHelp:
		msg = tgbotapi.NewMessage(query.Message.Chat.ID, "")
		msg.Text = `
You can configure a Google Sheets to store your expenses.
+ Step 1: Use the /gsheets command to configure a Google Sheets.
+ Step 2: Click the "Configure" button to provide the URL of your Google Sheets.
+ Step 3: Copy the URL of your Google Sheets and paste it in the chat.
+ Step 4: Share <b>Editing access</b> to the Google Sheets with the bot 
<b>housematee-gsheets@housematee.iam.gserviceaccount.com</b> (required).
<i> This is service account only, no one can access your Google Sheets except you.</i>
`
		e.endConversation(ctx, query.Message.Chat.ID)
	case gsheettypes.ActionUpdateCurrentPage:
		msg, _ = e.handleGSheetsUpdateCurrentPageCallback(ctx, query, subCommands)
	default:
		msg = tgbotapi.NewMessage(query.Message.Chat.ID, "")
		msg.Text = "Unfortunately, We have not supported this action yet."
		e.endConversation(ctx, query.Message.Chat.ID)
	}

	return msg, nil
}

func (e *Expensemate) handleGSheetsConfigureCallback(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "")
	msg.Text = `
Please provide the URL of your Google Sheets.
For example:
https://docs.google.com/spreadsheets/d/16jOEcyvHiHzW1GdRBvhHEadECojq0g3tzBT3a2MoLnI/edit?usp=sharing
`
	e.startConversation(
		ctx,
		query.Message.Chat.ID,
		tgtypes.BuildCallbackData(tgtypes.CommandGSheets, gsheettypes.ActionConfigure.String()),
	)
	return msg, nil
}

func (e *Expensemate) handleGSheetsConfigure(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	url := strings.TrimSpace(message.Text)
	if !httputils.IsValidGoogleSheetsURL(url) {
		msg.Text = "Invalid Google Sheets URL. Please provide a valid URL."
		return msg, nil
	}

	// find the user's Google Sheets URL
	// if the user has already configured a Google Sheets URL, update it
	// if the user has not configured a Google Sheets URL, save it

	userSheetMapping, err := e.getSpreadsheetMappingByUserID(ctx, message.From.ID)
	if err != nil {
		// Save the new one
		userSheetMapping = models.UserSheetMapping{
			UserID:          types.Id(message.Chat.ID),
			Username:        message.Chat.UserName,
			FullName:        message.Chat.FirstName + " " + message.Chat.LastName,
			SpreadSheetsURL: url,
			CreatedAt:       timeutils.GetCurrentDay(),
			UpdateAt:        timeutils.GetCurrentDay(),
			Status:          models.MappingStatusMapped,
		}
	} else {
		// Update the existing one
		userSheetMapping.SpreadSheetsURL = url
		userSheetMapping.UpdateAt = timeutils.GetCurrentDay()
	}

	err = e.upsertSpreadsheetMapping(ctx, userSheetMapping)
	if err != nil {
		msg.Text = "Failed to configure Google Sheets. Please try again."
		return msg, err
	}

	msg.Text = "Google Sheets is configured successfully. You can view it by clicking the /gsheets command."
	return msg, nil
}

func (e *Expensemate) handleGSheetsUpdateCurrentPageCallback(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	subCommands []string,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "")
	currentMonth := fmt.Sprintf(
		"%d_%02d",
		timeutils.GetCurrentDay().Year(),
		timeutils.GetCurrentDay().Month(),
	)
	switch {
	case len(subCommands) == 1: // gsheets:update_current_page
		msg.Text = fmt.Sprintf(
			`
Please select the sheet you want to update the current page <i>(Make sure you cloned the template from the bot.)</i> 
Or you can click the button below to update the current page to the current month => <b>%[1]s</b>.
`, currentMonth,
		)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf(
						"Update current page to %s",
						currentMonth,
					), fmt.Sprintf("gsheets:update_current_page:%s", currentMonth),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Other", "gsheets:update_current_page:other"),
			),
		)
	case len(subCommands) >= 2 && subCommands[1] == "other":
		msg, _ = e.showSheetList(ctx, query.Message)
	case len(subCommands) >= 2 && subCommands[1] != "":
		// update the current page to the selected month
		// the user has confirmed to update the current page to the selected month
		msg, _ = e.updateCurrentPageValue(ctx, query.Message, subCommands[1])
	}
	return msg, nil
}

func (e *Expensemate) showSheetList(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	mapping, err := e.getSpreadsheetMappingByUserID(ctx, message.Chat.ID)
	if err != nil {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}
	spreadsheetDocId := mapping.SpreadsheetDocId()
	sheets, err := gsheetclients.GetInstance().GetListOfSheets(ctx, spreadsheetDocId)
	if err != nil {
		msg.Text = "Failed to get the list of sheets."
		return msg, err
	}
	sheetsKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, sheet := range sheets {
		// filter sheets that have name format "YYYY_MM"
		if !gsheettypes.IsFormatValidSheetName(sheet.Properties.Title) {
			continue
		}

		sheetsKeyboard.InlineKeyboard = append(
			sheetsKeyboard.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					sheet.Properties.Title,
					fmt.Sprintf("gsheets:update_current_page:%s", sheet.Properties.Title),
				),
			),
		)
	}
	msg.ReplyMarkup = sheetsKeyboard
	msg.Text = "Please select the sheet you want to update the current page."
	return msg, nil
}

func (e *Expensemate) updateCurrentPageValue(
	ctx context.Context,
	message *tgbotapi.Message,
	newCurrentPageValue string,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	mapping, err := e.getSpreadsheetMappingByUserID(ctx, message.Chat.ID)
	if err != nil {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}
	spreadsheetDocId := mapping.SpreadsheetDocId()
	if err = e.updateCurrentPage(ctx, spreadsheetDocId, newCurrentPageValue); err != nil {
		msg.Text = "Failed to update the current page to the current month."
		return msg, err
	}
	msg.Text = fmt.Sprintf(
		"The current page is updated to <b>%s</b> successfully.",
		newCurrentPageValue,
	)
	e.endConversation(ctx, message.Chat.ID)
	return msg, nil
}

func (e *Expensemate) handleGSheetsUpdateCurrentPage(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	newCurrentPageValue := strings.TrimSpace(message.Text)
	return e.updateCurrentPageValue(ctx, message, newCurrentPageValue)
}
