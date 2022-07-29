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

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/internal/util"
	"github.com/illa-family/builder-backend/pkg/action"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/db"
	"github.com/illa-family/builder-backend/pkg/resource"
	"github.com/illa-family/builder-backend/pkg/state"
	filter "github.com/illa-family/builder-backend/pkg/websocket-filter"

	gws "github.com/gorilla/websocket"
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

// websocket client hub

var tssi *state.TreeStateServiceImpl
var kvssi *state.KVStateServiceImpl
var sssi *state.SetStateServiceImpl
var asi *app.AppServiceImpl
var rsi *resource.ResourceServiceImpl
var acsi *action.ActionServiceImpl

func initEnv() error {
	sugaredLogger := util.NewSugardLogger()
	dbConfig, err := db.GetConfig()
	if err != nil {
		return err
	}
	gormDB, err := db.NewDbConnection(dbConfig, sugaredLogger)
	if err != nil {
		return err
	}
	// init repo
	treestateRepositoryImpl := repository.NewTreeStateRepositoryImpl(sugaredLogger, gormDB)
	kvstateRepositoryImpl := repository.NewKVStateRepositoryImpl(sugaredLogger, gormDB)
	setstateRepositoryImpl := repository.NewSetStateRepositoryImpl(sugaredLogger, gormDB)
	appRepositoryImpl := repository.NewAppRepositoryImpl(sugaredLogger, gormDB)
	resourceRepositoryImpl := repository.NewResourceRepositoryImpl(sugaredLogger, gormDB)
	userRepositoryImpl := repository.NewUserRepositoryImpl(gormDB, sugaredLogger)
	actionRepositoryImpl := repository.NewActionRepositoryImpl(sugaredLogger, gormDB)
	// init service
	tssi = state.NewTreeStateServiceImpl(sugaredLogger, treestateRepositoryImpl)
	kvssi = state.NewKVStateServiceImpl(sugaredLogger, kvstateRepositoryImpl)
	sssi = state.NewSetStateServiceImpl(sugaredLogger, setstateRepositoryImpl)
	acsi = action.NewActionServiceImpl(sugaredLogger, actionRepositoryImpl)
	asi = app.NewAppServiceImpl(sugaredLogger, appRepositoryImpl, userRepositoryImpl, kvstateRepositoryImpl, treestateRepositoryImpl, setstateRepositoryImpl, actionRepositoryImpl)
	rsi = resource.NewResourceServiceImpl(sugaredLogger, resourceRepositoryImpl)
	return nil
}

var dashboardHub *ws.Hub
var appHub *ws.Hub

func InitHub(asi *app.AppServiceImpl, rsi *resource.ResourceServiceImpl, tssi *state.TreeStateServiceImpl, kvssi *state.KVStateServiceImpl, sssi *state.SetStateServiceImpl) {
	dashboardHub = ws.NewHub()
	dashboardHub.SetAppServiceImpl(asi)
	go filter.Run(dashboardHub)

	// init APP websocket hub
	appHub = ws.NewHub()
	appHub.SetResourceServiceImpl(rsi)
	appHub.SetTreeStateServiceImpl(tssi)
	appHub.SetKVStateServiceImpl(kvssi)
	appHub.SetSetStateServiceImpl(sssi)
	appHub.SetActionServiceImpl(acsi)
	go filter.Run(appHub)
}

// ServeWebsocket handle websocket requests from the peer.
func ServeWebsocket(hub *ws.Hub, w http.ResponseWriter, r *http.Request, instanceID string, appID int) {
	// init dashbroad websocket hub

	// @todo: this CheckOrigin method for debug only, remove it for release.
	upgrader := gws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Not a web socket connection: %s \n", err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}
	client := ws.NewClient(hub, conn, instanceID, appID)
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "websocket server serve address")
	flag.Parse()

	// init
	initEnv()
	InitHub(asi, rsi, tssi, kvssi, sssi)

	// listen and serve
	r := mux.NewRouter()
	// handle /status
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	// handle ws://{ip:port}/room/{instanceID}/dashboard
	r.HandleFunc("/room/{instanceID}/dashboard", func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instanceID"]
		log.Printf("[Connected] /room/%s/dashboard", instanceID)
		ServeWebsocket(dashboardHub, w, r, instanceID, ws.DEAULT_APP_ID)
	})
	// handle ws://{ip:port}/room/{instanceID}/app/{appID}
	r.HandleFunc("/room/{instanceID}/app/{appID}", func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instanceID"]
		appID, err := strconv.Atoi(mux.Vars(r)["appID"])
		if err != nil {
			appID = ws.DEAULT_APP_ID
		}
		log.Printf("[Connected] /room/%s/app/%d", instanceID, appID)
		ServeWebsocket(appHub, w, r, instanceID, appID)
	})
	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("[START] websocket service serve on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
