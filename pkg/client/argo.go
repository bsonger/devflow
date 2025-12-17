package client

import (
	"fmt"
	argoclient "github.com/argoproj/argo-cd/v3/pkg/client/clientset/versioned"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
)

var ArgoCdClient *argoclient.Clientset

// InitArgoCdClient 初始化 ArgoCD client
func InitArgoCdClient() error {
	var err error
	ArgoCdClient, err = argoclient.NewForConfig(model.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create argo cd client: %w", err)
	}
	logging.Logger.Info("argo cd client initialized")
	return nil
}
