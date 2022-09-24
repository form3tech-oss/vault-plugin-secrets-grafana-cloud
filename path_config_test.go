package secretsengine

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

const (
	key           = "1234567"
	configUrl     = "http://localhost:19090/"
	organisation  = "testorg"
	organisation1 = "testorg1"
)

func TestConfig(t *testing.T) {
	b, reqStorage := getTestBackend(t)

	t.Run("Test Configuration", func(t *testing.T) {
		t.Run("Create Configuration - pass", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          configUrl,
				"organisation": organisation,
			})
			assert.NoError(t, err)
		})

		t.Run("Create Configuration - empty key", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          "",
				"url":          configUrl,
				"organisation": organisation,
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - empty url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          "",
				"organisation": organisation,
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          "/addd",
				"organisation": organisation,
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid prometheus url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":            key,
				"url":            configUrl,
				"organisation":   organisation,
				"prometheus_url": "/p",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid loki url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          configUrl,
				"organisation": organisation,
				"loki_url":     "/l",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid tempo url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          configUrl,
				"organisation": organisation,
				"tempo_url":    "/t",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid alertmanager url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":              key,
				"url":              configUrl,
				"organisation":     organisation,
				"alertmanager_url": "/a",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - invalid graphite url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          configUrl,
				"organisation": organisation,
				"graphite_url": "/g",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration - empty organisation", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"key":          key,
				"url":          configUrl,
				"organisation": "",
			})
			assert.Error(t, err)
		})

		t.Run("Read Configuration - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"key":               key,
				"url":               configUrl,
				"organisation":      organisation,
				"user":              "",
				"prometheus_user":   "",
				"prometheus_url":    "",
				"loki_user":         "",
				"loki_url":          "",
				"tempo_user":        "",
				"tempo_url":         "",
				"alertmanager_user": "",
				"alertmanager_url":  "",
				"graphite_user":     "",
				"graphite_url":      "",
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (set users and urls) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"key":               key,
				"url":               "http://grafanacloud:19090",
				"organisation":      organisation1,
				"user":              "1",
				"prometheus_url":    "http://prometheus",
				"loki_user":         "2",
				"loki_url":          "http://loki",
				"tempo_user":        "3",
				"tempo_url":         "http://tempo",
				"alertmanager_user": "4",
				"alertmanager_url":  "http://alertmanager",
				"graphite_user":     "5",
				"graphite_url":      "http://graphite",
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration - invalid url", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"url": "abcde",
			})
			assert.Error(t, err)
		})

		t.Run("Read Updated Configuration (set users and urls) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"key":               key,
				"url":               "http://grafanacloud:19090",
				"organisation":      organisation1,
				"user":              "1",
				"prometheus_user":   "",
				"prometheus_url":    "http://prometheus",
				"loki_user":         "2",
				"loki_url":          "http://loki",
				"tempo_user":        "3",
				"tempo_url":         "http://tempo",
				"alertmanager_user": "4",
				"alertmanager_url":  "http://alertmanager",
				"graphite_user":     "5",
				"graphite_url":      "http://graphite",
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (set prometheus_user) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"key":             key,
				"url":             "http://grafanacloud:19090",
				"organisation":    organisation1,
				"prometheus_user": "6",
			})
			assert.NoError(t, err)
		})

		t.Run("Read Updated Configuration (set prometheus_user) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"key":               key,
				"url":               "http://grafanacloud:19090",
				"organisation":      organisation1,
				"user":              "6",
				"prometheus_user":   "6",
				"prometheus_url":    "http://prometheus",
				"loki_user":         "2",
				"loki_url":          "http://loki",
				"tempo_user":        "3",
				"tempo_url":         "http://tempo",
				"alertmanager_user": "4",
				"alertmanager_url":  "http://alertmanager",
				"graphite_user":     "5",
				"graphite_url":      "http://graphite",
			})
			assert.NoError(t, err)
		})

		t.Run("Delete Configuration - pass", func(t *testing.T) {
			err := testConfigDelete(b, reqStorage)
			assert.NoError(t, err)
		})
	})
}

func testConfigCreate(b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      configStoragePath,
		Data:      d,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigDelete(b logical.Backend, s logical.Storage) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      configStoragePath,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigUpdate(b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configStoragePath,
		Data:      d,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigRead(b logical.Backend, s logical.Storage, expected map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configStoragePath,
		Storage:   s,
	})

	if err != nil {
		return err
	}

	if resp == nil && expected == nil {
		return nil
	}

	if resp.IsError() {
		return resp.Error()
	}

	if len(expected) != len(resp.Data) {
		return fmt.Errorf("read data mismatch (expected %d values, got %d)", len(expected), len(resp.Data))
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]

		if !ok {
			return fmt.Errorf(`expected data["%s"] = %v but was not included in read output"`, k, expectedV)
		} else if expectedV != actualV {
			return fmt.Errorf(`expected data["%s"] = %v, instead got %v"`, k, expectedV, actualV)
		}
	}

	return nil
}
