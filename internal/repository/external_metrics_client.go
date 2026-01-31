package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

type UpdaterClient interface {
	ExternalUpdateMetrics(updateCounter int64, metrics map[string]float64) error
	ExternalUpdateJSONMetrics(updateCounter int64, metrics map[string]float64) error
}

type RestyUpdaterClient struct {
	httpClient *resty.Client
	host       string
}

func BuildRestyUpdaterMetric(host string) UpdaterClient {
	return RestyUpdaterClient{httpClient: resty.New(), host: host}
}

func (r RestyUpdaterClient) ExternalUpdateMetrics(updateCounter int64, metrics map[string]float64) error {
	for k, v := range metrics {
		urlModified := fmt.Sprintf("%s/update/%s/%s/%v", r.host, "gauge", k, v)
		resp, err := r.httpClient.R().SetHeader("Content-Type", "text/plain").Post(urlModified)
		if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
			log.Printf("err: error posting gauge metrics: %s, %v, %v, resp: %v", k, v, err, resp)
			return errors.New("error send gauge metrics")
		}
	}

	resp, err := r.httpClient.R().SetHeader("Content-Type", "text/plain").Post(
		fmt.Sprintf("%s/update/%s/%s/%v", r.host, "counter", "PollCount", updateCounter))
	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		log.Printf("err: error posting counter metrics: %v, resp: %v", err, resp)
		return errors.New("error send counter metrics")
	}
	return nil
}

func (r RestyUpdaterClient) ExternalUpdateJSONMetrics(updateCounter int64, metrics map[string]float64) error {
	urlModified := fmt.Sprintf("%s/update/", r.host)
	for k, v := range metrics {
		metric := models.Metrics{ID: k, MType: "gauge", Value: &v}
		jsonMetric, err := json.Marshal(metric)
		if err != nil {
			return errors.New("error convert to json gauge metrics")
		}
		resp, err := r.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetBody(jsonMetric).
			Post(urlModified)
		if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
			log.Printf("err: error posting gauge metrics: %s, %v, %v, resp: %v", k, v, err, resp)
			return errors.New("error send gauge metrics")
		}
	}

	metric := models.Metrics{ID: "PollCount", MType: "counter", Delta: &updateCounter}
	jsonMetric, err := json.Marshal(metric)
	if err != nil {
		return errors.New("error convert to json counter metrics")
	}
	resp, err := r.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonMetric).
		Post(urlModified)
	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		log.Printf("err: error posting counter metrics: %v, resp: %v", err, resp)
		return errors.New("error send counter metrics")
	}
	return nil
}
