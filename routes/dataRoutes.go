package routes

import (
	"fmt"
	"net/http"
	"telegram-spreadsheet-editor/errors"
	"telegram-spreadsheet-editor/model"
	"telegram-spreadsheet-editor/services"
)

type DataRoutes struct {
	DataService        services.IDataService
	SpreadsheetService services.ISpreadsheetService
	MessagingService   services.IMessagingService
	StorageService     services.IStorageService
}

func (r *DataRoutes) HandleMessage(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(resp, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	// get command
	command, err := r.MessagingService.GetCommandFromMessage(req.Body)
	if err != nil {
		switch err := err.(type) {
		case *errors.CommandError:
			// check if the user is not wanted
			if err.Unauthorized {
				r.MessagingService.SendTextMessage(err.ChatId, "Go away you prune head!")
				resp.WriteHeader(http.StatusOK)
				return
			}
			// otherwise just send back the response
			if err := r.MessagingService.SendTextMessage(err.ChatId, err.ResponseMessage); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		default:
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	}

	switch command.Type {
	case model.COMMAND_TYPE_PING:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "Pong"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_LIST:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "Listing... Hang tight..."); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		if err := r.MessagingService.SendEntryList(command.ChatId, entries); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_UPDATE:
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		if err := r.MessagingService.SendCategorySelectionKeyboard(command.ChatId, entries, "UPDATE"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN:
		if err := r.MessagingService.RemoveMarkupFromMessage(command.ChatId, command.MessageId); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("How much to we add to %s?", *command.UpdateData.Category)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_NUMERICAL_AMOUNT:
		// need to fetch the previous command
		prevCommand, err := r.StorageService.GetPreviousCommand(command.UserId)
		if err != nil {
			switch err := err.(type) {
			case *errors.StorageError:
				if err.Type == errors.STORAGE_ERROR_TYPE_NOT_FOUND {
					if err := r.MessagingService.SendTextMessage(command.ChatId, "Not sure what to do with that boyo. Type HELP."); err != nil {
						http.Error(resp, "", http.StatusFailedDependency)
						return
					}

					resp.WriteHeader(http.StatusOK)
					return
				}
			default:
				if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
					http.Error(resp, "", http.StatusFailedDependency)
					return
				}
				resp.WriteHeader(http.StatusOK)
				return
			}
		}

		if prevCommand.Type != model.COMMAND_TYPE_UPDATE_CATEGORY_CHOSEN {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Not sure what to do with that boyo. Type HELP."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}

			resp.WriteHeader(http.StatusOK)
			return
		}

		// merge commands
		fullCommand := model.MergeUpdateCommandWithFinancial(prevCommand, command)

		// ui feedback
		if err := r.MessagingService.SendTextMessage(command.ChatId, "On it, hang tight..."); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}

		// get sheet
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}

		// update sheet
		updated, newVal, err := r.SpreadsheetService.AddValueForCategory(sheet, *fullCommand.UpdateData.Category, *fullCommand.UpdateData.Value)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}

		// save sheet
		if err := r.DataService.WriteSpreadsheet(updated); err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}

		// done!
		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("Added £%.2f to %s. New total: %s", *fullCommand.UpdateData.Value, *fullCommand.UpdateData.Category, *newVal)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_READ:
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		if err := r.MessagingService.SendCategorySelectionKeyboard(command.ChatId, entries, "READ"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_READ_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(command.ChatId, command.MessageId)

		if err := r.MessagingService.SendTextMessage(command.ChatId, "On it, hang tight home slice..."); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}

		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		val, err := r.SpreadsheetService.ReadValueForCategory(sheet, command.ReadData.Category, false)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		// done!
		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("Current total for %s: %s", command.ReadData.Category, *val)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_DETAILS:
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		if err := r.MessagingService.SendCategorySelectionKeyboard(command.ChatId, entries, "DETAILS"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_DETAILS_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(command.ChatId, command.MessageId)

		if err := r.MessagingService.SendTextMessage(command.ChatId, "On it, hang tight home slice..."); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}

		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		val, err := r.SpreadsheetService.ReadValueForCategory(sheet, command.DetailsData.Category, true)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		// done!
		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("Details for for %s: %s", command.DetailsData.Category, *val)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_REMOVE:
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		entries, err := r.SpreadsheetService.ListCategoriesAndValues(sheet)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		if err := r.MessagingService.SendCategorySelectionKeyboard(command.ChatId, entries, "REMOVE"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_REMOVE_CATEGORY_CHOSEN:
		r.MessagingService.RemoveMarkupFromMessage(command.ChatId, command.MessageId)

		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("Removing last added value from %s", command.RemoveData.Category)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		// get sheet
		sheet, err := r.DataService.GetSpreadsheet()
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		// remove last value
		res, err := r.SpreadsheetService.RemoveLastValueForCategory(sheet, command.RemoveData.Category)
		if err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		// update sheet
		if err := r.DataService.WriteSpreadsheet(res.ModifiedSheet); err != nil {
			if err := r.MessagingService.SendTextMessage(command.ChatId, "Something went wrong..."); err != nil {
				http.Error(resp, "", http.StatusFailedDependency)
				return
			}
			resp.WriteHeader(http.StatusOK)
			return
		}
		// done!
		if err := r.MessagingService.SendTextMessage(command.ChatId, fmt.Sprintf("Removed %s from %s. Was %s and is now %s.", res.RemovedValue, command.RemoveData.Category, res.OldValue, res.NewValue)); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_HELP:
		helpText := `The following commands are available with this epic finance bot.
PING - Pong.
LIST - Lists all categories and their totals along with the running total.
UPDATE - Add a cost to a category.
READ - Read the value of a category.
DETAILS - Get all the costs of a category e.g. 2+4+6.
REMOVE - Delete the last added amount in a category.
HELP - Print this help list.`
		if err := r.MessagingService.SendTextMessage(command.ChatId, helpText); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_DORIS:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "\U0001F99B"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_BOOBS:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "They're great aren't they."); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	case model.COMMAND_TYPE_ALICE:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "Woooooo!"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		if err := r.MessagingService.SendTextMessage(command.ChatId, "\U0001F478"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		if err := r.MessagingService.SendTextMessage(command.ChatId, "\U00002728"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		if err := r.MessagingService.SendTextMessage(command.ChatId, "\u2764\ufe0f"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
		if err := r.MessagingService.SendTextMessage(command.ChatId, "\U0001F478 \U00002728 \u2764\ufe0f"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	}

	// update command for user for next call (THIS MUST GO LAST)
	r.StorageService.StoreCommand(command, command.UserId)

	resp.WriteHeader(http.StatusOK)
}
