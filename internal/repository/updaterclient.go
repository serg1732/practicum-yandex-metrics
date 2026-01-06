package repository

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

type IUpdaterClient interface {
	ExternalUpdateMetrics(updateCounter int64, metrics map[string]float64) error
}

type UpdaterClient struct {
	httpClient *http.Client
	url        string
}

func BuildUpdaterMetric(url string) IUpdaterClient {
	return UpdaterClient{httpClient: http.DefaultClient, url: url}
}

func (c UpdaterClient) ExternalUpdateMetrics(updateCounter int64, metrics map[string]float64) error {
	for k, v := range metrics {
		urlModified := fmt.Sprintf(c.url, "gauge", k, v)
		resp, err := c.httpClient.Post(urlModified, "Content-Type: text/plain", nil)
		if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
			log.Printf("err: error posting gauge metrics: %s, %v, %v, resp: %v", k, v, err, resp)
			return errors.New("error send gauge metrics")
		}
		errClose := resp.Body.Close()
		if errClose != nil {
			return errClose
		}
	}

	resp, err := c.httpClient.Post(fmt.Sprintf(c.url, "counter", "PollCount", updateCounter), "Content-Type: text/plain", nil)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		log.Printf("err: error posting counter metrics: %v, resp: %v", err, resp)
		return errors.New("error send counter metrics")
	}
	errClose := resp.Body.Close()
	if errClose != nil {
		return errClose
	}
	return nil
}
