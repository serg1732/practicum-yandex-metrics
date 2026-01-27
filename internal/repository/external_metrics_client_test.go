package repository

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExternalMetricsClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := BuildRestyUpdaterMetric(srv.URL)

	expectedPollCount := rand.Int64()
	err := client.ExternalUpdateMetrics(expectedPollCount, map[string]float64{
		"test-gauge": rand.Float64(),
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error send gauge metrics")
}

func TestExternalMetricsClientSuccess(t *testing.T) {
	expectedPollCount := rand.Int64()
	expectedKeyGauge := "test-gauge"
	expectedGauge := map[string]float64{expectedKeyGauge: rand.Float64()}
	expectedCounterURL := fmt.Sprintf("/update/counter/PollCount/%d", expectedPollCount)
	expectedGaugeURL := fmt.Sprintf("/update/gauge/%s/%v", expectedKeyGauge, expectedGauge[expectedKeyGauge])
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.URL.Path != expectedGaugeURL && r.URL.Path != expectedCounterURL {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := BuildRestyUpdaterMetric(srv.URL)
	err := client.ExternalUpdateMetrics(expectedPollCount, expectedGauge)
	assert.Nil(t, err)
}
