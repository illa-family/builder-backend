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

package filter

import (
	"github.com/illa-family/builder-backend/internal/repository"
	ws "github.com/illa-family/builder-backend/internal/websocket"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/state"
)

func SignalUpdateState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto *app.AppDto
	appDto.ConstructWithID(currentClient.APPID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil
	case ws.TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// update
			// construct update data
			var currentNode *state.TreeStateDto
			componentNode := repository.ConstructComponentNodeByMap(v)
			serializedComponent, err := componentNode.SerializationForDatabase()
			if err != nil {
				return err
			}
			currentNode.ConstructByMap(v) // set Name
			currentNode.ConstructWithContent(serializedComponent)
			currentNode.ConstructByApp(appDto) // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)

			// update
			if _, err := hub.TreeStateServiceImpl.UpdateTreeState(currentNode); err != nil {
				currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough

	case ws.TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough

	case ws.TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// update K-V State
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvStateDto *state.KVStateDto
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto)
			kvStateDto.ConstructWithType(stateType)

			// update
			if err := hub.KVStateServiceImpl.UpdateKVStateByKey(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dnsfu, err := repository.ConstructDisplayNameStateForUpdateByPayload(v)
			if err != nil {
				return err
			}
			// init state dto
			var beforeSetStateDto *state.SetStateDto
			var afterSetStateDto *state.SetStateDto
			beforeSetStateDto.ConstructWithDisplayNameForUpdate(dnsfu)
			beforeSetStateDto.ConstructWithType(stateType)
			beforeSetStateDto.ConstructByApp(appDto)
			beforeSetStateDto.ConstructWithEditVersion()
			afterSetStateDto.ConstructWithDisplayNameForUpdate(dnsfu)
			// update state
			if err := hub.SetStateServiceImpl.UpdateSetStateByValue(beforeSetStateDto, afterSetStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)

	return nil
}