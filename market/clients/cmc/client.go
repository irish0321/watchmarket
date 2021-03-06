package cmc

import (
	"github.com/trustwallet/blockatlas/pkg/blockatlas"
	"github.com/trustwallet/blockatlas/pkg/errors"
	"github.com/trustwallet/blockatlas/pkg/logger"
	"github.com/trustwallet/watchmarket/pkg/watchmarket"
	"net/url"
	"sync"
	"time"
)

const (
	minCachingDuration     = time.Minute * 5
	defaultCachingDuration = time.Minute * 10
)

type Client struct {
	apiKey string
	blockatlas.Request
	mu              sync.Mutex
	CachingDuration time.Duration
}

func NewClient(api, key string, cachingDuration time.Duration) *Client {
	c := Client{
		Request: blockatlas.InitClient(api),
		apiKey:  key,
	}
	c.Headers["X-CMC_PRO_API_KEY"] = key
	if cachingDuration < minCachingDuration {
		logger.Warn("Setting CMC caching to default", logger.Params{"caching": defaultCachingDuration})
		c.CachingDuration = defaultCachingDuration
	} else {
		c.CachingDuration = cachingDuration
	}

	return &c
}

func (c *Client) GetData() (prices CoinPrices, err error) {
	c.mu.Lock()
	request := blockatlas.Request{
		BaseUrl:      c.BaseUrl,
		Headers:      c.Headers,
		HttpClient:   blockatlas.DefaultClient,
		ErrorHandler: blockatlas.DefaultErrorHandler,
	}
	err = request.GetWithCache(&prices, "v1/cryptocurrency/listings/latest", url.Values{"limit": {"5000"}, "convert": {watchmarket.DefaultCurrency}}, c.CachingDuration)
	c.mu.Unlock()
	return
}

func GetCmcMap(mapApi string) (CmcMapping, error) {
	var results CmcSlice
	request := blockatlas.Request{
		BaseUrl:      mapApi,
		HttpClient:   blockatlas.DefaultClient,
		ErrorHandler: blockatlas.DefaultErrorHandler,
	}
	err := request.GetWithCache(&results, "mapping.json", nil, time.Hour*1)
	if err != nil {
		return nil, errors.E(err).PushToSentry()
	}
	return results.cmcToCoinMap(), nil
}

func GetCoinMap(mapApi string) (CoinMapping, error) {
	var results CmcSlice
	request := blockatlas.Request{
		BaseUrl:      mapApi,
		HttpClient:   blockatlas.DefaultClient,
		ErrorHandler: blockatlas.DefaultErrorHandler,
	}
	err := request.GetWithCache(&results, "mapping.json", nil, time.Hour*1)
	if err != nil {
		return nil, errors.E(err).PushToSentry()
	}
	return results.coinToCmcMap(), nil
}
