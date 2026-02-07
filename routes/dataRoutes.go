package routes

import (
	"fmt"
	"telegram-spreadsheet-editor/errors"
	"telegram-spreadsheet-editor/model"
	"telegram-spreadsheet-editor/services"

	"go.uber.org/zap"
)

type DataRoutes struct {
	DataService        services.IDataService
	SpreadsheetService services.ISpreadsheetService
	MessagingService   services.IMessagingService
	StorageService     services.IStorageService
}

func (r *DataRoutes) HandleMessage(message *model.Message) {
	zap.L().Info("Starting message handle")

	// get command
	// to handle multiple inputs we need a master message service that calls the sub ones to get the info it needs
	command, err := r.MessagingService.GetCommandFromMessage(message)
	if err != nil {
		switch err := err.(type) {
		case *errors.CommandError:
			// check if the user is not wanted
			if err.Unauthorized {
				r.MessagingService.SendTextMessage(message, err.ChatId, "Go away you prune head!")
				return
			}
			// otherwise just send back the response
			r.MessagingService.SendTextMessage(message, err.ChatId, err.ResponseMessage)
			return
		default:
			return
		}
	}

	zap.L().Info("Handling message", zap.Uint8("type", command.Type))

	switch command.Type {
	case model.COMMAND_TYPE_PING:
		r.MessagingService.SendTextMessage(message, command.ChatId, "Pong")
	case model.COMMAND_TYPE_LIST:
		r.MessagingService.SendTextMessage(message, command.ChatId, "Listing... Hang tight...")
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		r.MessagingService.SendEntryList(message, command.ChatId, entries)
	case model.COMMAND_TYPE_UPDATE:
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		r.MessagingService.SendCategorySelectionKeyboard(message, command.ChatId, entries, "UPDATE")
	case model.COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(message, command.ChatId, command.MessageId)
		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("How much to we add to %s?", *command.UpdateData.Category))
	case model.COMMAND_TYPE_NUMERICAL_AMOUNT:
		// need to fetch the previous command
		prevCommand, err := r.StorageService.GetPreviousCommand(command.UserId)
		if err != nil {
			switch err := err.(type) {
			case *errors.StorageError:
				if err.Type == errors.STORAGE_ERROR_TYPE_NOT_FOUND {
					r.MessagingService.SendTextMessage(message, command.ChatId, "Not sure what to do with that boyo. Type HELP.")
					return
				}
			default:
				r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
				return
			}
		}

		if prevCommand.Type != model.COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Not sure what to do with that boyo. Type HELP.")
			return
		}

		// merge commands
		fullCommand := model.MergeUpdateCommandWithFinancial(prevCommand, command)

		// ui feedback
		r.MessagingService.SendTextMessage(message, command.ChatId, "On it, hang tight...")

		// get sheet
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}

		// update sheet
		updated, newVal, err := r.SpreadsheetService.AddValueForCategory(sheet, *fullCommand.UpdateData.Category, *fullCommand.UpdateData.Value)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}

		// save sheet
		if err := r.DataService.WriteSpreadsheet(source, updated); err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}

		// done!
		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("Added £%.2f to %s. New total: %s", *fullCommand.UpdateData.Value, *fullCommand.UpdateData.Category, *newVal))
	case model.COMMAND_TYPE_READ:
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		r.MessagingService.SendCategorySelectionKeyboard(message, command.ChatId, entries, "READ")
	case model.COMMAND_TYPE_READ_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(message, command.ChatId, command.MessageId)
		r.MessagingService.SendTextMessage(message, command.ChatId, "On it, hang tight home slice...")

		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		val, err := r.SpreadsheetService.ReadValueForCategory(sheet, command.ReadData.Category, false)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		// done!
		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("Current total for %s: %s", command.ReadData.Category, *val))
	case model.COMMAND_TYPE_DETAILS:
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		r.MessagingService.SendCategorySelectionKeyboard(message, command.ChatId, entries, "DETAILS")
	case model.COMMAND_TYPE_DETAILS_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(message, command.ChatId, command.MessageId)

		r.MessagingService.SendTextMessage(message, command.ChatId, "On it, hang tight home slice...")

		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		val, err := r.SpreadsheetService.ReadValueForCategory(sheet, command.DetailsData.Category, true)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		// done!
		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("Details for for %s: %s", command.DetailsData.Category, *val))
	case model.COMMAND_TYPE_REMOVE:
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		r.MessagingService.SendCategorySelectionKeyboard(message, command.ChatId, entries, "REMOVE")
	case model.COMMAND_TYPE_REMOVE_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(message, command.ChatId, command.MessageId)

		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("Removing last added value from %s", command.RemoveData.Category))
		// get sheet
		source := r.getSpreadsheetSource(message.UserName)
		sheet, err := r.DataService.GetSpreadsheet(source)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		// remove last value
		res, err := r.SpreadsheetService.RemoveLastValueForCategory(sheet, command.RemoveData.Category)
		if err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		// update sheet
		if err := r.DataService.WriteSpreadsheet(source, res.ModifiedSheet); err != nil {
			r.MessagingService.SendTextMessage(message, command.ChatId, "Something went wrong...")
			return
		}
		// done!
		r.MessagingService.SendTextMessage(message, command.ChatId, fmt.Sprintf("Removed %s from %s. Was %s and is now %s.", res.RemovedValue, command.RemoveData.Category, res.OldValue, res.NewValue))
	case model.COMMAND_TYPE_HELP:
		helpText := `The following commands are available with this epic finance bot.
PING - Pong.
LIST - Lists all categories and their totals along with the running total.
UPDATE - Add a cost to a category.
READ - Read the value of a category.
DETAILS - Get all the costs of a category e.g. 2+4+6.
REMOVE - Delete the last added amount in a category.
HELP - Print this help list.`
		r.MessagingService.SendTextMessage(message, command.ChatId, helpText)
	case model.COMMAND_TYPE_DORIS:
		r.MessagingService.SendTextMessage(message, command.ChatId, "\U0001F99B")
	case model.COMMAND_TYPE_BOOBS:
		r.MessagingService.SendTextMessage(message, command.ChatId, "They're great aren't they.")
	case model.COMMAND_TYPE_ALICE:
		r.MessagingService.SendTextMessage(message, command.ChatId, "Woooooo!")
		r.MessagingService.SendTextMessage(message, command.ChatId, "\U0001F478")
		r.MessagingService.SendTextMessage(message, command.ChatId, "\U00002728")
		r.MessagingService.SendTextMessage(message, command.ChatId, "\u2764\ufe0f")
		r.MessagingService.SendTextMessage(message, command.ChatId, "\U0001F478 \U00002728 \u2764\ufe0f")
	}

	// update command for user for next call (THIS MUST GO LAST)
	r.StorageService.StoreCommand(command, command.UserId)
}

// private
func (r *DataRoutes) getSpreadsheetSource(userName string) model.SpreadsheetSource {
	config := model.GetConfig()
	for _, u := range config.Users {
		if u.Name == userName {
			return u.SpreadsheetSource
		}
	}

	return nil
}
