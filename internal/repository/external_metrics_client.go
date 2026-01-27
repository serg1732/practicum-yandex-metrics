package repository

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type UpdaterClient interface {
	ExternalUpdateMetrics(updateCounter int64, metrics map[string]float64) error
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
