package main

import (
	"context"
	"fmt"
	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	"github.com/bsonger/devflow/pkg/config"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/otel"
	"github.com/bsonger/devflow/pkg/router"
	"github.com/bsonger/devflow/pkg/tekton"
	"go.uber.org/zap"
)

// @title			DevFlow CD Platform API
// @version		1.0
// @description	DevFlow CD 平台 REST API
// @termsOfService	http://devflow.example.com/terms/
// @contact.name	DevFlow Team
// @contact.url	http://devflow.example.com
// @contact.email	devflow@example.com
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @schemes		http https
func main() {
	if err := config.Load(); err != nil {
		panic(err)
	}
	logging.Init()

	ctx := context.Background()
	otel.Init(config.C.Otel.Endpoint, config.C.Otel.ServiceName)

	_, err := db.InitMongo(ctx, config.C.Mongo.URI, config.C.Mongo.DBName, logging.Logger)
	if err != nil {
		logging.Logger.Fatal("mongo init failed", zap.Error(err))
	}

	err = tekton.InitTektonClient()
	if err != nil {
		logging.Logger.Fatal("tekton init failed", zap.Error(err))
	}

	logging.Logger.Info("server start")
	r := router.NewRouter()

	port := config.C.Server.Port
	logging.Logger.Info("starting server", zap.Int("port", port))
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		logging.Logger.Fatal("failed to run server", zap.Error(err))
	}
}
