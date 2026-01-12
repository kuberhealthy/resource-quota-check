package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

// main loads configuration and runs the resource quota check.
func main() {
	// Parse configuration from environment variables.
	cfg, err := parseConfig()
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Apply debug settings after parsing.
	applyDebugSettings(cfg)

	// Create a timeout context for the check.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckTimeLimit)
	defer cancel()

	// Start listening for interrupts.
	go listenForInterrupts(cancel)

	// Create a Kubernetes client.
	client, err := createKubeClient(cfg.KubeConfigFile)
	if err != nil {
		reportFailureAndExit(fmt.Errorf("failed to create a kubernetes client: %w", err))
		return
	}
	log.Infoln("Kubernetes client created.")

	// Catch panics and report failures.
	defer handlePanic()

	// Run the resource quota check.
	rqErrors, err := runResourceQuotaCheck(ctx, cfg, client)
	if err != nil {
		handleRunError(err)
		return
	}

	// Report failure when any errors are found.
	if len(rqErrors) != 0 {
		log.Infoln("This check created", len(rqErrors), "errors and warnings.")
		log.Debugln("Errors and warnings:")
		for _, rqErr := range rqErrors {
			log.Debugln(rqErr)
		}
		reportErr := checkclient.ReportFailure(rqErrors)
		if reportErr != nil {
			log.Fatalln("error reporting failures to kuberhealthy:", reportErr.Error())
		}
		return
	}

	// Report success when no errors are found.
	log.Infoln("Reporting success to kuberhealthy.")
	reportErr := checkclient.ReportSuccess()
	if reportErr != nil {
		log.Fatalln("error reporting success to kuberhealthy:", reportErr.Error())
	}
}

// handleRunError reports runtime errors to Kuberhealthy.
func handleRunError(err error) {
	// Report a timeout error when the context deadline is exceeded.
	if err == context.DeadlineExceeded {
		reportFailureAndExit(fmt.Errorf("Check took too long and timed out."))
		return
	}

	// Report generic failures to Kuberhealthy.
	reportFailureAndExit(err)
}

// listenForInterrupts listens for OS interrupts and cancels the check.
func listenForInterrupts(cancel context.CancelFunc) {
	// Relay incoming OS interrupt signals.
	signalChan := make(chan os.Signal, 3)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	sig := <-signalChan
	log.Infoln("Received an interrupt signal from the signal channel.")
	log.Debugln("Signal received was:", sig.String())

	// Cancel the context to stop work.
	log.Debugln("Cancelling context.")
	cancel()

	// Clean up and exit after a second signal or timeout.
	log.Infoln("Shutting down.")

	select {
	case sig = <-signalChan:
		log.Warnln("Received a second interrupt signal from the signal channel.")
		log.Debugln("Signal received was:", sig.String())
	case <-time.After(30 * time.Second):
		log.Infoln("Clean up took too long to complete and timed out.")
	}

	os.Exit(0)
}

// handlePanic reports panics as failures.
func handlePanic() {
	// Recover from panics during the check.
	recovered := recover()
	if recovered == nil {
		return
	}

	// Report the panic to Kuberhealthy.
	log.Infoln("Recovered panic:", recovered)
	reportErr := checkclient.ReportFailure([]string{fmt.Sprint(recovered)})
	if reportErr != nil {
		log.Fatalln("error reporting failure to kuberhealthy:", reportErr.Error())
	}
}
