package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultThreshold is the default usage alert threshold.
	defaultThreshold = 0.9
	// defaultCheckTimeLimit is the default timeout for the check.
	defaultCheckTimeLimit = 5 * time.Minute
)

// CheckConfig stores configuration for the resource quota check.
type CheckConfig struct {
	// KubeConfigFile is the optional kubeconfig path.
	KubeConfigFile string
	// Blacklist namespaces to skip.
	Blacklist []string
	// Whitelist namespaces to include.
	Whitelist []string
	// Threshold is the CPU/memory usage threshold.
	Threshold float64
	// CheckTimeLimit is the maximum runtime for the check.
	CheckTimeLimit time.Duration
	// Debug enables debug logging.
	Debug bool
}

// parseConfig reads environment variables and builds a CheckConfig.
func parseConfig() (*CheckConfig, error) {
	// Parse debug settings.
	debug, err := parseDebugSetting()
	if err != nil {
		return nil, err
	}

	// Parse blacklist and whitelist namespaces.
	blacklist := []string{}
	whitelist := []string{}
	blacklistEnv := os.Getenv("BLACKLIST")
	if len(blacklistEnv) != 0 {
		blacklist = strings.Split(blacklistEnv, ",")
		log.Infoln("Parsed BLACKLIST:", blacklist)
	}
	whitelistEnv := os.Getenv("WHITELIST")
	if len(whitelistEnv) != 0 {
		whitelist = strings.Split(whitelistEnv, ",")
		log.Infoln("Parsed WHITELIST:", whitelist)
	}

	// Parse memory and CPU thresholds.
	threshold := defaultThreshold
	thresholdEnv := os.Getenv("THRESHOLD")
	if len(thresholdEnv) != 0 {
		parsedThreshold, parseErr := strconv.ParseFloat(thresholdEnv, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("error occurred attempting to parse THRESHOLD: %w", parseErr)
		}
		threshold = parsedThreshold
		log.Infoln("Parsed THRESHOLD:", threshold)
	}
	if threshold > 0.99 {
		log.Infoln("Given THRESHOLD is greater than 0.99, setting to default of", defaultThreshold)
		threshold = defaultThreshold
	}
	if threshold <= 0 {
		log.Infoln("Threshold is less than or equal to 0, setting to default of", defaultThreshold)
		threshold = defaultThreshold
	}
	log.Infoln("Usage threshold set to:", threshold)

	// Set check time limit to default.
	checkTimeLimit := defaultCheckTimeLimit

	// Override using the Kuberhealthy deadline when available.
	timeDeadline, err := checkclient.GetDeadline()
	if err != nil {
		log.Infoln("There was an issue getting the check deadline:", err.Error())
	}
	checkTimeLimit = timeDeadline.Sub(time.Now().Add(time.Second * 5))
	log.Infoln("Check time limit set to:", checkTimeLimit)

	// Assemble configuration.
	cfg := &CheckConfig{}
	cfg.KubeConfigFile = os.Getenv("KUBECONFIG")
	cfg.Blacklist = blacklist
	cfg.Whitelist = whitelist
	cfg.Threshold = threshold
	cfg.CheckTimeLimit = checkTimeLimit
	cfg.Debug = debug

	return cfg, nil
}

// parseDebugSetting parses the DEBUG environment variable.
func parseDebugSetting() (bool, error) {
	// Default to disabled debug logging.
	debug := false

	// Parse DEBUG when provided.
	debugEnv := os.Getenv("DEBUG")
	if len(debugEnv) != 0 {
		parsedDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			return false, fmt.Errorf("failed to parse DEBUG environment variable: %w", err)
		}
		debug = parsedDebug
	}

	return debug, nil
}

// applyDebugSettings updates logrus based on the debug flag.
func applyDebugSettings(cfg *CheckConfig) {
	// Enable debug logging when requested.
	if !cfg.Debug {
		return
	}

	// Apply logrus debug settings.
	log.Infoln("Debug logging enabled.")
	log.SetLevel(log.DebugLevel)
	log.Debugln(os.Args)
}
