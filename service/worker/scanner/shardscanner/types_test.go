package shardscanner

import (
	"context"
	"testing"
)

func TestGetFixerContextPanicsWhenNonActivityContextProvided(t *testing.T) {
	defer func() { recover() }()
	GetFixerContext(context.Background())
	t.Errorf("should have panicked")
}

func TestGetScannerContextPanicsWhenNonActivityContextProvided(t *testing.T) {
	defer func() { recover() }()
	GetScannerContext(context.Background())
	t.Errorf("should have panicked")
}
