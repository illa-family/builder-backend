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

package ws

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

// message protocol from client in json:
//
// {
//     "signal":number,
//     "option":number(work as int32 bit),
//     "payload":string
// }

// for message
const SIGNAL_PING = 0
const SIGNAL_ENTER = 1
const SIGNAL_LEAVE = 2
const SIGNAL_CREATE_STATE = 3
const SIGNAL_DELETE_STATE = 4
const SIGNAL_UPDATE_STATE = 5
const SIGNAL_MOVE_STATE = 6
const SIGNAL_CREATE_OR_UPDATE_STATE = 7
const SIGNAL_ONLY_BROADCAST = 8
const SIGNAL_PUT_STATE = 9

const OPTION_BROADCAST_ROOM = 1 // 00000000000000000000000000000001; // use as signed int32 in typescript

const TARGET_NOTNING = 0            // placeholder for nothing
const TARGET_COMPONENTS = 1         // ComponentsState
const TARGET_DEPENDENCIES = 2       // DependenciesState
const TARGET_DRAG_SHADOW = 3        // DragShadowState
const TARGET_DOTTED_LINE_SQUARE = 4 // DottedLineSquareState
const TARGET_DISPLAY_NAME = 5       // DisplayNameState
const TARGET_APPS = 6               // only for broadcast
const TARGET_RESOURCE = 7           // only for broadcast
const TARGET_ACTION = 8             // ActionState

// for broadcast rewrite
const BROADCAST_TYPE_SUFFIX = "/remote"

type Broadcast struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Message struct {
	ClientID  uuid.UUID     `json:"clientID"`
	Signal    int           `json:"signal"`
	APPID     int           `json:"appID"` // also as APP ID
	Option    int           `json:"option"`
	Target    int           `json:"target"`
	Payload   []interface{} `json:"payload"`
	Broadcast *Broadcast    `json:"broadcast"`
}

func NewMessage(clientID uuid.UUID, appID int, rawMessage []byte) (*Message, error) {
	// init Action
	var message Message
	if err := json.Unmarshal(rawMessage, &message); err != nil {
		return nil, err
	}
	message.ClientID = clientID
	message.APPID = appID
	return &message, nil
}

func (m *Message) RewriteBroadcast() {
	m.Broadcast.Type = m.Broadcast.Type + BROADCAST_TYPE_SUFFIX
}
