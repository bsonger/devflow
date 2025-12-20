package main

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow-common/client/logging"
	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	"github.com/bsonger/devflow/pkg/config"
	"github.com/bsonger/devflow/pkg/router"
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
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	err = config.InitConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	r := router.NewRouter()

	port := cfg.Server.Port
	logging.Logger.Info("server start")
	logging.Logger.Info("starting server", zap.Int("port", port))
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		logging.Logger.Fatal("failed to run server", zap.Error(err))
	}
}
