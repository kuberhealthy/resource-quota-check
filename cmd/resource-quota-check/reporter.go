package main

import (
	"os"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

// reportFailureAndExit reports failure to Kuberhealthy and exits.
func reportFailureAndExit(err error) {
	// Report the failure to Kuberhealthy.
	reportErr := checkclient.ReportFailure([]string{err.Error()})
	if reportErr != nil {
		log.Fatalln("error reporting failure to kuberhealthy:", reportErr.Error())
	}

	// Exit after reporting failure.
	os.Exit(0)
}
