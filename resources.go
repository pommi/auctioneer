package auctioneer

import (
	"errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/rep"
)

type TaskStartRequest struct {
	rep.Task
}

func NewTaskStartRequest(task rep.Task) TaskStartRequest {
	return TaskStartRequest{task}
}

func NewTaskStartRequestFromModel(taskGuid, domain string, taskDef *models.TaskDefinition) TaskStartRequest {
	volumeMounts := []string{}
	for _, volumeMount := range taskDef.VolumeMounts {
		volumeMounts = append(volumeMounts, volumeMount.Driver)
	}
	return TaskStartRequest{
		rep.NewTask(
			taskGuid,
			domain,
			rep.NewPlacementConstraint([]string{}, volumeMounts, taskDef.RootFs),
			rep.NewResource(taskDef.MemoryMb, taskDef.DiskMb),
		),
	}
}

func (t *TaskStartRequest) Validate() error {
	switch {
	case t.TaskGuid == "":
		return errors.New("task guid is empty")
	case t.Resource.Empty():
		return errors.New("resources cannot be empty")
	case t.PlacementConstraint.Empty():
		return errors.New("placement constraint cannot be empty")
	default:
		return nil
	}
}

type LRPStartRequest struct {
	ProcessGuid string `json:"process_guid"`
	Domain      string `json:"domain"`
	Indices     []int  `json:"indices"`
	rep.PlacementConstraint
	rep.Resource
}

func NewLRPStartRequest(processGuid, domain string, indices []int, pl rep.PlacementConstraint, res rep.Resource) LRPStartRequest {
	return LRPStartRequest{
		ProcessGuid:         processGuid,
		Domain:              domain,
		Indices:             indices,
		PlacementConstraint: pl,
		Resource:            res,
	}
}

func NewLRPStartRequestFromModel(d *models.DesiredLRP, indices ...int) LRPStartRequest {
	volumeDrivers := []string{}
	for _, volumeMount := range d.VolumeMounts {
		volumeDrivers = append(volumeDrivers, volumeMount.Driver)
	}

	return NewLRPStartRequest(
		d.ProcessGuid,
		d.Domain,
		indices,
		rep.NewPlacementConstraint(d.PlacementTags, volumeDrivers, d.RootFs),
		rep.NewResource(d.MemoryMb, d.DiskMb),
	)
}

func NewLRPStartRequestFromSchedulingInfo(s *models.DesiredLRPSchedulingInfo, indices ...int) LRPStartRequest {
	return NewLRPStartRequest(
		s.ProcessGuid,
		s.Domain,
		indices,
		rep.NewPlacementConstraint(s.PlacementTags, s.VolumePlacement.DriverNames, s.RootFs),
		rep.NewResource(s.MemoryMb, s.DiskMb),
	)
}

func (lrpstart *LRPStartRequest) Validate() error {
	switch {
	case lrpstart.ProcessGuid == "":
		return errors.New("proccess guid is empty")
	case lrpstart.Domain == "":
		return errors.New("domain is empty")
	case len(lrpstart.Indices) == 0:
		return errors.New("indices must not be empty")
	case lrpstart.Resource.Empty():
		return errors.New("resources cannot be empty")
	case lrpstart.PlacementConstraint.Empty():
		return errors.New("placement constraint cannot be empty")
	default:
		return nil
	}
}
