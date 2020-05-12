package findata

import (
	"encoding/json"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"io/ioutil"
	"net/http"
	"net/url"
)

type (
	CmcClient struct {
		apiKey string
		proxy  string
	}
)

func (c *CmcClient) GetDetails() ([]AssetDetail, error) {
	ids, err := c.getAssetIds()
	if err != nil {
		return nil, err
	}

	var res []AssetDetail
	for i := 0; i < len(ids); i += 100 {
		length := 100
		if i+length > len(ids) {
			length = len(ids) - i
		}
		details, err := c.getDetails(ids[i : i+length])
		if err != nil {
			return nil, err
		}
		res = append(res, details...)
	}
	return res, nil
}

// api document
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyMap
func (c *CmcClient) getAssetIds() ([]int, error) {
	type (
		cmcId struct {
			Id int `json:"id"`
		}

		cmcIds struct {
			Data []cmcId `json:"data"`
		}
	)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/map", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("listing_status", "active") // active, inactive, untracked
	q.Add("start", "1")               // start with id 1
	q.Add("sort", "id")               // sort by id
	q.Add("limit", "5000")            // max limit 5000

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", c.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ids := cmcIds{}
	if err := json.Unmarshal(respBody, &ids); err != nil {
		return nil, err
	}
	var res []int
	for _, v := range ids.Data {
		res = append(res, v.Id)
	}
	return res, nil
}

func (c CmcClient) getDetails(ids []int) ([]AssetDetail, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/info", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("id", gnum.JoinFormatInt(",", ids...))

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", c.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Status != "200 OK" {
		return nil, gerror.New(resp.Status)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type (
		cmcDetail struct {
			Urls struct {
				Website      []string `json:"website"`
				TechnicalDoc []string `json:"technical_doc"`
				Twitter      []string `json:"twitter"`
				Reddit       []string `json:"reddit"`
				MessageBoard []string `json:"message_board"`
				Announcement []string `json:"announcement"`
				Chat         []string `json:"chat"`
				Explorer     []string `json:"explorer"`
				SourceCode   []string `json:"source_code"`
			} `json:"urls"`
			Logo   string `json:"logo"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			Slug   string `json:"slug"`
		}

		cmcDetails struct {
			Data map[string]cmcDetail `json:"data"`
		}
	)
	getSocialMedias := func(detail cmcDetail) []string {
		var res []string
		res = append(res, detail.Urls.Reddit...)
		res = append(res, detail.Urls.SourceCode...)
		res = append(res, detail.Urls.Chat...)
		res = append(res, detail.Urls.Announcement...)
		res = append(res, detail.Urls.Twitter...)
		res = append(res, detail.Urls.MessageBoard...)
		return res
	}

	cds := cmcDetails{}
	if err := json.Unmarshal(respBody, &cds); err != nil {
		return nil, err
	}
	var res []AssetDetail
	for _, v := range cds.Data {
		item := AssetDetail{
			Asset:       fintypes.NewCoin(v.Slug, v.Symbol, fintypes.CoinMarketCap),
			Logo:        v.Logo,
			Websites:    v.Urls.Website,
			Explorers:   v.Urls.Explorer,
			TechDocs:    v.Urls.TechnicalDoc,
			SocialMedia: getSocialMedias(v),
		}
		res = append(res, item)
	}

	return res, nil
}
