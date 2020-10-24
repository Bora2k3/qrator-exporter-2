package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements prometheus.Collector and little bit a custom fields
type Collector struct {
	url                     string
	id                      string
	authToken               string
	up                      prometheus.Gauge
	failedScrapes           prometheus.Counter
	totalScrapes            prometheus.Counter
	failedDomainScrapes     prometheus.Counter
	failedStatisticsScrapes prometheus.CounterVec
	failedJSONDecode        prometheus.CounterVec
	sync.Mutex

	BandwidthTraffic prometheus.GaugeVec
	Packets          prometheus.GaugeVec
	Blacklist        prometheus.GaugeVec
	HTTPRequests     prometheus.GaugeVec
	HTTPResponses    prometheus.GaugeVec
	HTTPErrors       prometheus.GaugeVec
}

type requestData struct {
	ID     int      `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

// ResponseData returned HTTP data from Qrator API
type ResponseData struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// Domain data of single domain record
type Domain struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	IP        []string    `json:"ip"`
	IPJson    interface{} `json:"ip_json"`
	QratorIP  string      `json:"qratorIp"`
	IsService bool        `json:"isService"`
	Ports     interface{} `json:"ports"`
}

// StatisticsCurrentIP contains statistics for single domain record
type StatisticsCurrentIP struct {
	Time      int `json:"time"`
	Bandwidth struct {
		Input  float64 `json:"input"`
		Passed float64 `json:"passed"`
		Output float64 `json:"output"`
	} `json:"bandwidth"`
	Packets struct {
		Input  float64 `json:"input"`
		Passed float64 `json:"passed"`
		Output float64 `json:"output"`
	} `json:"packets"`
	Blacklist struct {
		Qrator float64 `json:"qrator"`
		API    float64 `json:"api"`
		WAF    float64 `json:"waf"`
	} `json:"blacklist"`
}

// StatisticsCurrentHTTP contains HTTP statistics for single domain record
type StatisticsCurrentHTTP struct {
	Time      int     `json:"time"`
	Requests  float64 `json:"requests"`
	Responses struct {
		Duration0000to0200 float64 `json:"0000_0200" mapstructure:"0000_0200"`
		Duration0200to0500 float64 `json:"0200_0500" mapstructure:"0200_0500"`
		Duration0500to0700 float64 `json:"0500_0700" mapstructure:"0500_0700"`
		Duration0700to1000 float64 `json:"0700_1000" mapstructure:"0700_1000"`
		Duration1000to1500 float64 `json:"1000_1500" mapstructure:"1000_1500"`
		Duration1500to2000 float64 `json:"1500_2000" mapstructure:"1500_2000"`
		Duration2000to5000 float64 `json:"2000_5000" mapstructure:"2000_5000"`
		Duration5000toInf  float64 `json:"5000_inf" mapstructure:"5000_inf"`
	} `json:"responses"`
	Errors struct {
		Total   float64 `json:"total"`
		HTTP500 float64 `json:"500" mapstructure:"500"`
		HTTP501 float64 `json:"501" mapstructure:"501"`
		HTTP502 float64 `json:"502" mapstructure:"502"`
		HTTP503 float64 `json:"503" mapstructure:"503"`
		HTTP504 float64 `json:"504" mapstructure:"504"`
	} `json:"errors"`
}
