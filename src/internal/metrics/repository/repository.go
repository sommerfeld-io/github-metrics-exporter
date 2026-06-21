// Package repository provides the repository accessibility gauge metric.
package repository

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// Accessible is the gauge that reports whether each discovered repository is reachable.
// Label dimensions: owner, repo. Value: 1 = accessible, 0 = inaccessible.
// It is initialized by calling Init.
var Accessible *prometheus.GaugeVec

// ErrNotInitialized is returned by Set when Init has not been called.
var ErrNotInitialized = errors.New("repository.Set: Init must be called first")

// Init creates the Accessible gauge using the given metric prefix.
// It returns an error if prefix is empty.
func Init(prefix string) error {
	if prefix == "" {
		return errors.New("repository.Init: prefix must not be empty")
	}
	Accessible = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefix + "repository_accessible",
			Help: "1 if the repository is accessible; 0 if a 403 or 404 was returned.",
		},
		[]string{"owner", "repo"},
	)
	return nil
}

// Set records the accessibility status for a single repository.
// It returns ErrNotInitialized if Init has not been called.
func Set(owner, name string, accessible bool) error {
	if Accessible == nil {
		return ErrNotInitialized
	}
	var value float64
	if accessible {
		value = 1.0
	}
	Accessible.WithLabelValues(owner, name).Set(value)
	return nil
}
