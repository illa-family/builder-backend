// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package action

import (
	"errors"
	"time"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/app"
	"go.uber.org/zap"
)

var type_array = [8]string{"transformer", "restapi", "graphql", "redis", "mysql", "mariadb", "postgresql", "mongodb"}
var type_map = map[string]int{
	"transformer": 0,
	"restapi":     1,
	"graphql":     2,
	"redis":       3,
	"mysql":       4,
	"mariadb":     5,
	"postgresql":  6,
	"mongodb":     7,
}

type ActionService interface {
	CreateAction(action ActionDto) (ActionDto, error)
	DeleteAction(id int) error
	UpdateAction(action ActionDto) (ActionDto, error)
	GetAction(id int) (ActionDto, error)
	FindActionsByAppVersion(app, version int) ([]ActionDto, error)
	RunAction(action ActionDto) (interface{}, error)
	ValidateActionOptions(actionType string, options map[string]interface{}) error
}

type ActionDto struct {
	ID          int                    `json:"actionId"`
	App         int                    `json:"-"`
	Version     int                    `json:"-"`
	Resource    int                    `json:"resourceId,omitempty"`
	DisplayName string                 `json:"displayName" validate:"required"`
	Type        string                 `json:"actionType" validate:"oneof=transformer restapi graphql redis mysql mariadb postgresql mongodb"`
	Template    map[string]interface{} `json:"content" validate:"required"`
	Transformer map[string]interface{} `json:"transformer" validate:"required"`
	TriggerMode string                 `json:"triggerMode" validate:"oneof=manually automate"`
	CreatedAt   time.Time              `json:"createdAt,omitempty"`
	CreatedBy   int                    `json:"createdBy,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy   int                    `json:"updatedBy,omitempty"`
}

type ActionServiceImpl struct {
	logger             *zap.SugaredLogger
	actionRepository   repository.ActionRepository
	resourceRepository repository.ResourceRepository
}

func NewActionDto() *ActionDto {
	return &ActionDto{}
}

func (ad *ActionDto) ConstructByMap(data interface{}) error {
	udata, ok := data.(map[string]interface{})
	if !ok {
		err := errors.New("ActionDto construct by map failed, please check your input payload syntax.")
		return err
	}
	for k, v := range udata {
		switch k {
		case "displayName":
			ad.DisplayName, _ = v.(string)
		case "actionType":
			ad.Type, _ = v.(string)
		case "transformer":
			ad.Transformer, _ = v.(map[string]interface{})
		case "triggerMode":
			ad.TriggerMode, _ = v.(string)
		case "resourceId":
			appf, _ := v.(float64)
			ad.Resource = int(appf)
		case "content":
			ad.Template, _ = v.(map[string]interface{})
		}
	}
	return nil
}

func (ad *ActionDto) ConstructIDByMap(data interface{}) error {
	udata, ok := data.(map[string]interface{})
	if !ok {
		err := errors.New("ActionDto construct by map failed, please check your input payload syntax.")
		return err
	}
	for k, v := range udata {
		switch k {
		case "actionId":
			ad.ID, _ = v.(int)
		}
	}
	return nil
}

func (ad *ActionDto) ConstructWithDisplayNameForDelete(displayNameInterface interface{}) error {
	dnis, ok := displayNameInterface.(string)
	if !ok {
		err := errors.New("ConstructWithDisplayNameForDelete() can not resolve displayName.")
		return err
	}
	ad.DisplayName = dnis
	return nil
}

func (ad *ActionDto) ConstructByApp(app *app.AppDto) {
	ad.App = app.ID
}

func NewActionServiceImpl(logger *zap.SugaredLogger, actionRepository repository.ActionRepository,
	resourceRepository repository.ResourceRepository) *ActionServiceImpl {
	return &ActionServiceImpl{
		logger:             logger,
		actionRepository:   actionRepository,
		resourceRepository: resourceRepository,
	}
}

func (impl *ActionServiceImpl) CreateAction(action ActionDto) (ActionDto, error) {
	// TODO: guarantee `action` DisplayName unique
	id, err := impl.actionRepository.Create(&repository.Action{
		ID:          action.ID,
		App:         action.App,
		Version:     action.Version,
		Resource:    action.Resource,
		Name:        action.DisplayName,
		Type:        type_map[action.Type],
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		CreatedAt:   action.CreatedAt,
		CreatedBy:   action.CreatedBy,
		UpdatedAt:   action.UpdatedAt,
		UpdatedBy:   action.UpdatedBy,
	})
	if err != nil {
		return ActionDto{}, err
	}
	action.ID = id

	return action, nil
}

func (impl *ActionServiceImpl) DeleteAction(id int) error {
	if err := impl.actionRepository.Delete(id); err != nil {
		return err
	}
	return nil
}

func (impl *ActionServiceImpl) UpdateAction(action ActionDto) (ActionDto, error) {
	// TODO: guarantee `action` DisplayName unique
	if err := impl.actionRepository.Update(&repository.Action{
		ID:          action.ID,
		Resource:    action.Resource,
		Name:        action.DisplayName,
		Type:        type_map[action.Type],
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		UpdatedAt:   action.UpdatedAt,
		UpdatedBy:   action.UpdatedBy,
	}); err != nil {
		return ActionDto{}, err
	}
	return action, nil
}

func (impl *ActionServiceImpl) GetAction(id int) (ActionDto, error) {
	res, err := impl.actionRepository.RetrieveByID(id)
	if err != nil {
		return ActionDto{}, err
	}
	resDto := ActionDto{
		ID:          res.ID,
		Resource:    res.Resource,
		DisplayName: res.Name,
		Type:        type_array[res.Type],
		TriggerMode: res.TriggerMode,
		Transformer: res.Transformer,
		Template:    res.Template,
		CreatedBy:   res.CreatedBy,
		CreatedAt:   res.CreatedAt,
		UpdatedBy:   res.UpdatedBy,
		UpdatedAt:   res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *ActionServiceImpl) FindActionsByAppVersion(app, version int) ([]ActionDto, error) {
	res, err := impl.actionRepository.RetrieveActionsByAppVersion(app, version)
	if err != nil {
		return nil, err
	}

	resDtoSlice := make([]ActionDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, ActionDto{
			ID:          value.ID,
			Resource:    value.Resource,
			DisplayName: value.Name,
			Type:        type_array[value.Type],
			TriggerMode: value.TriggerMode,
			Transformer: value.Transformer,
			Template:    value.Template,
			CreatedBy:   value.CreatedBy,
			CreatedAt:   value.CreatedAt,
			UpdatedBy:   value.UpdatedBy,
			UpdatedAt:   value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *ActionServiceImpl) RunAction(action ActionDto) (interface{}, error) {
	if action.Resource == 0 {
		return nil, errors.New("resource is required")
	}
	rsc, err := impl.resourceRepository.RetrieveByID(action.Resource)
	if err != nil {
		return nil, err
	}
	actionFactory := Factory{Type: action.Type}
	actionAssemblyLine := actionFactory.Build()
	if actionAssemblyLine == nil {
		return nil, errors.New("invalid ActionType:: unsupported type")
	}
	if _, err := actionAssemblyLine.ValidateResourceOptions(rsc.Options); err != nil {
		return nil, errors.New("invalid resource content")
	}
	if _, err := actionAssemblyLine.ValidateActionOptions(action.Template); err != nil {
		return nil, errors.New("invalid action content")
	}
	res, err := actionAssemblyLine.Run(rsc.Options, action.Template)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ActionServiceImpl) ValidateActionOptions(actionType string, options map[string]interface{}) error {
	if actionType == TRANSFORMER_ACTION {
		return nil
	}
	actionFactory := Factory{Type: actionType}
	actionAssemblyLine := actionFactory.Build()
	if actionAssemblyLine == nil {
		return errors.New("invalid ActionType:: unsupported type")
	}
	if _, err := actionAssemblyLine.ValidateActionOptions(options); err != nil {
		return errors.New("invalid action content")
	}
	return nil
}
