package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// runResourceQuotaCheck inspects quotas across namespaces and returns errors.
func runResourceQuotaCheck(ctx context.Context, cfg *CheckConfig, client *kubernetes.Clientset) ([]string, error) {
	// List all namespaces in the cluster.
	allNamespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error occurred listing namespaces from the cluster: %w", err)
	}

	// Collect errors and warnings.
	errors := make([]string, 0)

	// Iterate through namespaces with filtering.
	for _, ns := range allNamespaces.Items {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if shouldSkipNamespace(ns.GetName(), cfg) {
			continue
		}

		quotaErrors := examineResourceQuotasForNamespace(ctx, ns.GetName(), cfg.Threshold, client)
		errors = append(errors, quotaErrors...)
	}

	log.Infoln("No errors or warnings were created during this check!")
	return errors, nil
}

// shouldSkipNamespace applies blacklist/whitelist filters.
func shouldSkipNamespace(namespace string, cfg *CheckConfig) bool {
	// Prioritize blacklist over the whitelist.
	if len(cfg.Blacklist) > 0 {
		if contains(namespace, cfg.Blacklist) {
			log.Infoln("Skipping", namespace, "namespace (Blacklist).")
			return true
		}
	}

	if len(cfg.Whitelist) > 0 {
		if !contains(namespace, cfg.Whitelist) {
			log.Infoln("Skipping", namespace, "namespace (Whitelist).")
			return true
		}
	}

	return false
}

// examineResourceQuotasForNamespace inspects quota usage and returns errors.
func examineResourceQuotasForNamespace(ctx context.Context, namespace string, threshold float64, client *kubernetes.Clientset) []string {
	// List resource quotas in the namespace.
	log.Infoln("Looking at resource quotas for", namespace, "namespace.")
	quotas, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []string{fmt.Sprintf("error occurred listing resource quotas for %s namespace %v", namespace, err)}
	}

	// Evaluate quota usage against the threshold.
	errors := make([]string, 0)
	for _, rq := range quotas.Items {
		limits := rq.Status.Hard
		status := rq.Status.Used
		percentCPUUsed := float64(status.Cpu().MilliValue()) / float64(limits.Cpu().MilliValue())
		percentMemoryUsed := float64(status.Memory().MilliValue()) / float64(limits.Memory().MilliValue())
		log.Debugln("Current used for", namespace, "CPU:", status.Cpu().MilliValue(), "Memory:", status.Memory().MilliValue())
		log.Debugln("Limits for", namespace, "CPU:", limits.Cpu().MilliValue(), "Memory:", limits.Memory().MilliValue())
		if percentCPUUsed >= threshold {
			err := fmt.Errorf("cpu for %s namespace has reached threshold of %4.2f: USED: %d LIMIT: %d PERCENT_USED: %6.3f",
				namespace, threshold, status.Cpu().MilliValue(), limits.Cpu().MilliValue(), percentCPUUsed)
			errors = append(errors, err.Error())
		}
		if percentMemoryUsed >= threshold {
			err := fmt.Errorf("memory for %s namespace has reached threshold of %4.2f: USED: %d LIMIT: %d PERCENT_USED: %6.3f",
				namespace, threshold, status.Memory().MilliValue(), limits.Memory().MilliValue(), percentMemoryUsed)
			errors = append(errors, err.Error())
		}
	}

	return errors
}

// contains checks if a slice contains a value.
func contains(value string, list []string) bool {
	// Iterate through the list and check for matches.
	for _, item := range list {
		if value == item {
			return true
		}
	}

	return false
}
