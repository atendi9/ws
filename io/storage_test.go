package io

import (
	"math/rand"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestStorageSizeMultiplication(t *testing.T) {
	// Initialize a new random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	t.Run("Validate KB with random multiplier", func(t *testing.T) {
		randomValue := StorageSize(rng.Intn(1000) + 1)
		
		expected := randomValue * 1024
		result := randomValue * KB
		
		assert.Equal(t, expected, result)
	})

	t.Run("Validate MB with random multiplier", func(t *testing.T) {
		randomValue := StorageSize(rng.Intn(1000) + 1)
		
		expected := randomValue * 1024 * 1024
		result := randomValue * MB
		
		assert.Equal(t, expected, result)
	})

	t.Run("Validate GB with random multiplier", func(t *testing.T) {
		randomValue := StorageSize(rng.Intn(1000) + 1)
		
		expected := randomValue * 1024 * 1024 * 1024
		result := randomValue * GB
		
		assert.Equal(t, expected, result)
	})

	t.Run("Validate TB with random multiplier", func(t *testing.T) {
		randomValue := StorageSize(rng.Intn(1000) + 1)
		
		expected := randomValue * 1024 * 1024 * 1024 * 1024
		result := randomValue * TB
		
		assert.Equal(t, expected, result)
	})

	t.Run("Validate PB with random multiplier", func(t *testing.T) {
		randomValue := StorageSize(rng.Intn(1000) + 1)
		
		expected := randomValue * 1024 * 1024 * 1024 * 1024 * 1024
		result := randomValue * PB
		
		assert.Equal(t, expected, result)
	})
}