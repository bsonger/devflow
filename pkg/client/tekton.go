package client

import (
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
)

var TektonClient *tektonclient.Clientset

func InitTektonClient() error {
	var err error
	TektonClient, err = tektonclient.NewForConfig(model.KubeConfig)
	if err != nil {
		return err
	}
	logging.Logger.Info("init tekton client")
	return nil
}
