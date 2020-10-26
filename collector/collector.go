package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace     = "qrator"
	apiRequestURL = "https://api.qrator.net/request"
	onlineStatus  = "online"
)

// NewCollector creates a new Collector type
func NewCollector(id, token string) (*Collector, error) {
	collector := Collector{
		url:       apiRequestURL,
		id:        id,
		authToken: token,
	}

	ping, err := HTTPRequest("client", "ping", collector.id, collector.authToken)
	if err != nil {
		return nil, err
	}
	defer ping.Body.Close()

	response, err := DecodeResponse(ping)
	if err != nil {
		return nil, fmt.Errorf("Request method 'ping' got status '%v'", err)
	}

	if response.Result != "pong" {
		return nil, fmt.Errorf("Request method 'ping' got invalid status from Qrator API")
	}

	collector.metrics()

	return &collector, nil
}

// HTTPRequest makes http request to Qrator API
func HTTPRequest(methodClass, method, id, authToken string) (*http.Response, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", apiRequestURL, methodClass, id)
	data := requestData{
		ID:     1,
		Method: method,
	}

	body, _ := json.Marshal(data)
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))

	if err != nil {
		return nil, fmt.Errorf("Cannot create new request: %v", err)
	}
	defer request.Body.Close()

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Qrator-Auth", authToken)

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("Cannot make new request: %v", err)
	}

	return response, nil
}

// DecodeResponse decode json response from Qrator API
func DecodeResponse(httpResponse *http.Response) (*ResponseData, error) {
	var decode ResponseData
	err := json.NewDecoder(httpResponse.Body).Decode(&decode)

	if err != nil {
		return nil, fmt.Errorf("Got error while decoding json. %v", err)
	}

	if decode.Error != "" {
		return nil, fmt.Errorf("%s", decode.Error)
	}

	return &decode, nil
}

func (c *Collector) domainsList() ([]Domain, error) {
	var domains []Domain

	response, err := HTTPRequest("client", "domains_get", c.id, c.authToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	decodeResult, err := DecodeResponse(response)
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(decodeResult.Result, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func (c *Collector) onlineDomains() ([]Domain, error) {
	domains, err := c.domainsList()
	if err != nil {
		return nil, err
	}

	var result []Domain

	for _, domain := range domains {
		if domain.Status == onlineStatus && !domain.IsService {
			result = append(result, domain)
		}
	}

	return result, nil
}

// Describe implements prometheus.Collector (2)
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *Collector) registerJSONDecodeError(domainName, apiMethod string, ch chan<- prometheus.Metric) {
	c.failedStatisticsScrapes.WithLabelValues(domainName, apiMethod).Inc()
	ch <- c.failedStatisticsScrapes.WithLabelValues(domainName, apiMethod)
}

func (c *Collector) scripe(ch chan<- prometheus.Metric) error {
	processError := make(chan error)
	processDone := make(chan bool)

	c.totalScrapes.Inc()

	onlineDomains, err := c.onlineDomains()
	if err != nil {
		c.failedDomainScrapes.Inc()
		return err
	}

	waitGroup := &sync.WaitGroup{}

	for _, domain := range onlineDomains {
		waitGroup.Add(2)

		go func(domain Domain, ch chan<- prometheus.Metric, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()

			var (
				decodeStatistics StatisticsCurrentHTTP
				apiMethod        = "statistics_current_http"
			)

			response, err := HTTPRequest("domain", apiMethod, strconv.Itoa(domain.ID), c.authToken)
			if err != nil {
				c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod).Inc()
				ch <- c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod)
				processError <- err
			}
			defer response.Body.Close()

			statistics, err := DecodeResponse(response)
			if err != nil {
				c.registerJSONDecodeError(domain.Name, apiMethod, ch)
				processError <- err
			}

			err = mapstructure.Decode(statistics.Result, &decodeStatistics)
			if err != nil {
				c.registerJSONDecodeError(domain.Name, apiMethod, ch)
				processError <- err
			}

			c.HTTPRequests.WithLabelValues(domain.Name, apiMethod).Set(decodeStatistics.Requests)

			c.HTTPResponses.WithLabelValues(domain.Name, "0.0-0.2s", apiMethod).Set(decodeStatistics.Responses.Duration0000to0200)
			c.HTTPResponses.WithLabelValues(domain.Name, "0.2-0.5s", apiMethod).Set(decodeStatistics.Responses.Duration0200to0500)
			c.HTTPResponses.WithLabelValues(domain.Name, "0.5-0.7s", apiMethod).Set(decodeStatistics.Responses.Duration0500to0700)
			c.HTTPResponses.WithLabelValues(domain.Name, "0.7-1.0s", apiMethod).Set(decodeStatistics.Responses.Duration0700to1000)
			c.HTTPResponses.WithLabelValues(domain.Name, "1.0-1.5s", apiMethod).Set(decodeStatistics.Responses.Duration1000to1500)
			c.HTTPResponses.WithLabelValues(domain.Name, "1.5-2.0s", apiMethod).Set(decodeStatistics.Responses.Duration1500to2000)
			c.HTTPResponses.WithLabelValues(domain.Name, "2.0-5.0s", apiMethod).Set(decodeStatistics.Responses.Duration2000to5000)
			c.HTTPResponses.WithLabelValues(domain.Name, "over 5s", apiMethod).Set(decodeStatistics.Responses.Duration5000toInf)

			c.HTTPErrors.WithLabelValues(domain.Name, "Total", apiMethod).Set(decodeStatistics.Errors.Total)
			c.HTTPErrors.WithLabelValues(domain.Name, "500", apiMethod).Set(decodeStatistics.Errors.HTTP500)
			c.HTTPErrors.WithLabelValues(domain.Name, "501", apiMethod).Set(decodeStatistics.Errors.HTTP501)
			c.HTTPErrors.WithLabelValues(domain.Name, "502", apiMethod).Set(decodeStatistics.Errors.HTTP502)
			c.HTTPErrors.WithLabelValues(domain.Name, "503", apiMethod).Set(decodeStatistics.Errors.HTTP503)
			c.HTTPErrors.WithLabelValues(domain.Name, "504", apiMethod).Set(decodeStatistics.Errors.HTTP504)

			ch <- c.HTTPRequests.WithLabelValues(domain.Name, apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "0.0-0.2s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "0.2-0.5s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "0.5-0.7s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "0.7-1.0s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "1.0-1.5s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "1.5-2.0s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "2.0-5.0s", apiMethod)
			ch <- c.HTTPResponses.WithLabelValues(domain.Name, "over 5s", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "Total", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "500", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "501", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "502", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "503", apiMethod)
			ch <- c.HTTPErrors.WithLabelValues(domain.Name, "504", apiMethod)

			ch <- c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod)
			ch <- c.failedJSONDecode.WithLabelValues(domain.Name, apiMethod)
		}(domain, ch, waitGroup)

		go func(domain Domain, ch chan<- prometheus.Metric, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()

			var (
				decodeStatistics StatisticsCurrentIP
				apiMethod        = "statistics_current_ip"
			)

			response, err := HTTPRequest("domain", apiMethod, strconv.Itoa(domain.ID), c.authToken)
			if err != nil {
				c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod).Inc()
				ch <- c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod)
				processError <- err
			}
			defer response.Body.Close()

			statistics, err := DecodeResponse(response)
			if err != nil {
				c.registerJSONDecodeError(domain.Name, apiMethod, ch)
				processError <- err
			}

			err = mapstructure.Decode(statistics.Result, &decodeStatistics)
			if err != nil {
				c.registerJSONDecodeError(domain.Name, apiMethod, ch)
				processError <- err
			}

			c.BandwidthTraffic.WithLabelValues(domain.Name, "input", apiMethod).Set(decodeStatistics.Bandwidth.Input)
			c.BandwidthTraffic.WithLabelValues(domain.Name, "passed", apiMethod).Set(decodeStatistics.Bandwidth.Passed)
			c.BandwidthTraffic.WithLabelValues(domain.Name, "output", apiMethod).Set(decodeStatistics.Bandwidth.Output)

			c.Packets.WithLabelValues(domain.Name, "input", apiMethod).Set(decodeStatistics.Packets.Input)
			c.Packets.WithLabelValues(domain.Name, "passed", apiMethod).Set(decodeStatistics.Packets.Passed)
			c.Packets.WithLabelValues(domain.Name, "output", apiMethod).Set(decodeStatistics.Packets.Output)

			c.Blacklist.WithLabelValues(domain.Name, "qrator", apiMethod).Set(decodeStatistics.Blacklist.Qrator)
			c.Blacklist.WithLabelValues(domain.Name, "api", apiMethod).Set(decodeStatistics.Blacklist.API)
			c.Blacklist.WithLabelValues(domain.Name, "waf", apiMethod).Set(decodeStatistics.Blacklist.WAF)

			ch <- c.BandwidthTraffic.WithLabelValues(domain.Name, "input", apiMethod)
			ch <- c.BandwidthTraffic.WithLabelValues(domain.Name, "passed", apiMethod)
			ch <- c.BandwidthTraffic.WithLabelValues(domain.Name, "output", apiMethod)
			ch <- c.Packets.WithLabelValues(domain.Name, "input", apiMethod)
			ch <- c.Packets.WithLabelValues(domain.Name, "passed", apiMethod)
			ch <- c.Packets.WithLabelValues(domain.Name, "output", apiMethod)
			ch <- c.Blacklist.WithLabelValues(domain.Name, "qrator", apiMethod)
			ch <- c.Blacklist.WithLabelValues(domain.Name, "api", apiMethod)
			ch <- c.Blacklist.WithLabelValues(domain.Name, "waf", apiMethod)

			ch <- c.failedStatisticsScrapes.WithLabelValues(domain.Name, apiMethod)
			ch <- c.failedJSONDecode.WithLabelValues(domain.Name, apiMethod)
		}(domain, ch, waitGroup)
	}

	go func() {
		waitGroup.Wait()
		close(processDone)
	}()

	select {
	case <-processDone:
		break
	case err := <-processError:
		close(processError)
		return err
	}

	return nil
}

// Collect implements prometheus.Collector (1)
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	err := c.scripe(ch)
	if err != nil {
		c.failedScrapes.Inc()
		c.up.Set(0)
		log.Println(err)
	} else {
		c.up.Set(1)
	}

	ch <- c.up
	ch <- c.totalScrapes
	ch <- c.failedScrapes
	ch <- c.failedDomainScrapes
}
