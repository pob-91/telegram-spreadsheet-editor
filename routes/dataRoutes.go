package routes

import (
	"encoding/json"
	"net/http"
	"nextcloud-spreadsheet-editor/errors"
	"nextcloud-spreadsheet-editor/model"
	"nextcloud-spreadsheet-editor/services"
)

type DataRoutes struct {
	DataService        services.IDataService
	SpreadsheetService services.ISpreadsheetService
	MessagingService   services.IMessagingService
}

type AddValueRequest struct {
	Category string  `json:"category"`
	Value    float32 `json:"value"`
}

func (r *DataRoutes) AddValueForCategory(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request AddValueRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	spreadsheet, err := r.DataService.GetSpreadsheet()
	if err != nil {
		resp.WriteHeader(http.StatusFailedDependency)
		return
	}

	modified, err := r.SpreadsheetService.AddValueForCategory(spreadsheet, request.Category, request.Value)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.DataService.WriteSpreadsheet(modified); err != nil {
		resp.WriteHeader(http.StatusFailedDependency)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
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
				http.Error(resp, "", http.StatusForbidden)
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
	case model.COMMAND_TYPE_LIST:
		if err := r.MessagingService.SendTextMessage(command.ChatId, "Listing categories"); err != nil {
			http.Error(resp, "", http.StatusFailedDependency)
			return
		}
	}

	resp.WriteHeader(http.StatusOK)
}
