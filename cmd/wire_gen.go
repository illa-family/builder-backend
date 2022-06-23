// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/api/resthandler"
	"github.com/illa-family/builder-backend/api/router"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/internal/util"
	"github.com/illa-family/builder-backend/pkg/action"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/db"
	"github.com/illa-family/builder-backend/pkg/resource"
)

// Injectors from wire.go:

func Initialize() (*Server, error) {
	config, err := GetAppConfig()
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	sugaredLogger := util.NewSugardLogger()
	dbConfig, err := db.GetConfig()
	if err != nil {
		return nil, err
	}
	gormDB, err := db.NewDbConnection(dbConfig, sugaredLogger)
	if err != nil {
		return nil, err
	}
	appRepositoryImpl := repository.NewAppRepositoryImpl(sugaredLogger, gormDB)
	appServiceImpl := app.NewAppServiceImpl(sugaredLogger, appRepositoryImpl)
	appRestHandlerImpl := resthandler.NewAppRestHandlerImpl(sugaredLogger, appServiceImpl)
	appRouterImpl := router.NewAppRouterImpl(appRestHandlerImpl)
	actionRepositoryImpl := repository.NewActionRepositoryImpl(sugaredLogger, gormDB)
	actionServiceImpl := action.NewActionServiceImpl(sugaredLogger, actionRepositoryImpl)
	actionRestHandlerImpl := resthandler.NewActionRestHandlerImpl(sugaredLogger, actionServiceImpl)
	actionRouterImpl := router.NewActionRouterImpl(actionRestHandlerImpl)
	resourceRepositoryImpl := repository.NewResourceRepositoryImpl(sugaredLogger, gormDB)
	resourceServiceImpl := resource.NewResourceServiceImpl(sugaredLogger, resourceRepositoryImpl)
	resourceRestHandlerImpl := resthandler.NewResourceRestHandlerImpl(sugaredLogger, resourceServiceImpl)
	resourceRouterImpl := router.NewResourceRouterImpl(resourceRestHandlerImpl)
	restRouter := router.NewRESTRouter(sugaredLogger, appRouterImpl, actionRouterImpl, resourceRouterImpl)
	server := NewServer(config, engine, restRouter, sugaredLogger)
	return server, nil
}