package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type VolumeAuctionHandler struct {
	runner auctiontypes.AuctionRunner
}

func NewVolumeAuctionHandler(runner auctiontypes.AuctionRunner) *VolumeAuctionHandler {
	return &VolumeAuctionHandler{
		runner: runner,
	}
}

func (*VolumeAuctionHandler) logSession(logger lager.Logger) lager.Logger {
	return logger.Session("volume-auction-handler")
}

func (h *VolumeAuctionHandler) Create(w http.ResponseWriter, r *http.Request, logger lager.Logger) {
	logger = h.logSession(logger).Session("create-volume-set")
	logger.Info("started")

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed-to-read-request-body", err)
		writeInternalErrorJSONResponse(w, err)
		return
	}

	starts := []models.VolumeStartRequest{}
	err = json.Unmarshal(payload, &starts)
	if err != nil {
		logger.Error("malformed-json", err)
		writeInvalidJSONResponse(w, err)
		return
	}

	validStarts := make([]models.VolumeStartRequest, 0, len(starts))
	for _, start := range starts {
		if err := start.Validate(); err == nil {
			validStarts = append(validStarts, start)
		} else {
			logger.Error("start-validate-failed", err, lager.Data{"volume-start": start})
		}
	}

	h.runner.ScheduleVolumesForAuctions(validStarts)
	logger.Info("submitted")
	writeStatusAcceptedResponse(w)
}
