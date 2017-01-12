package handlers

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/auction/auctiontypes"
	"code.cloudfoundry.org/lager"
)

type ReportHandler struct {
	runner auctiontypes.AuctionRunner
}

func NewReportHandler(runner auctiontypes.AuctionRunner) *ReportHandler {
	return &ReportHandler{
		runner: runner,
	}
}

func (*ReportHandler) logSession(logger lager.Logger) lager.Logger {
	return logger.Session("lrp-auction-handler")
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request, logger lager.Logger) {
	logger = h.logSession(logger).Session("report")

	j := struct {
		TaskCount  int32  `json:"taskCount"`
		LrpCount   int32  `json:"lrpCount"`
		Identifier string `json:"Identifier"`
	}{TaskCount: taskAuctionCallCount, LrpCount: lrpAuctionCallCount, Identifier: h.runner.Identifier()}
	jstring, err := json.Marshal(j)
	if err != nil {
		logger.Fatal("Error in marshalling", err)
	}
	logger.Info("report-response", lager.Data{"value": string(jstring)})
	writeJSONResponse(w, http.StatusOK, j)
}
