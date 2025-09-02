package routes

import (
	"encoding/json"
	"net/http"
	"nextcloud-spreadsheet-editor/services"
)

type DataRoutes struct {
	DataService        services.IDataService
	SpreadsheetService services.ISpreadsheetService
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
