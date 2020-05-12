package findata

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/net/ghtml"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"strings"
	"time"
)

type (
	CgkClient struct {
		proxy string
	}

	cgkCoinName struct {
		Name   string `json:"id"`
		Symbol string `json:"symbol"`
	}

	cgkCoinDetail struct {
		Error string `json:"error"`

		Name              string  `json:"id"`
		Symbol            string  `json:"symbol"`
		MaxSupply         float64 `json:"MaxSupply,omitempty"`
		TotalSupply       float64 `json:"TotalSupply,omitempty"`
		CirculatingSupply float64 `json:"CirculatingSupply,omitempty"`

		Links struct {
			Homepage                    []string `json:"homepage"`
			BlockchainSite              []string `json:"blockchain_site"`
			OfficialForumUrl            []string `json:"official_forum_url"`
			ChatUrl                     []string `json:"chat_url"`
			AnnouncementUrl             []string `json:"announcement_url"`
			TwitterScreenName           string   `json:"twitter_screen_name"`
			FacebookUsername            string   `json:"facebook_username"`
			BitcointalkThreadIdentifier int      `json:"bitcointalk_thread_identifier"`
			TelegramChannelIdentifier   string   `json:"telegram_channel_identifier"`
			SubRedditUrl                string   `json:"subreddit_url"`
			ReposUrl                    struct {
				Github    []string `json:"github"`
				Bitbucket []string `json:"bitbucket"`
			} `json:"repos_url"`
		} `json:"links"`

		Image struct {
			Thumb string `json:"thumb"`
			Small string `json:"small"`
			Large string `json:"large"`
		} `json:"image"`

		GenesisDate string `json:"genesis_date"`
		/*
			IcoData struct{
				IcoStartDate time.Time `json:"ico_start_date"`
				IcoEndDate time.Time `json:"ico_end_date"`
				BasePublicSaleAmount float64 `json:"base_public_sale_amount"` // "1"
				QuotePublicSaleCurrency string `json:"quote_public_sale_currency"` // "BTC" / "ETH"
				QuotePublicSaleAmount float64 `json:"quote_public_sale_amount"` // "0.00012"
				TotalRaisedCurrency string `json:"total_raised_currency"` // "USD"
				TotalRaised string `json:"total_raised"` // "36000000"
			} `json:"ico_data"`*/

		CoingeckoScore      float64 `json:"coingecko_score"`
		DeveloperScore      float64 `json:"developer_score"`
		CommunityScore      float64 `json:"community_score"`
		LiquidityScore      float64 `json:"liquidity_score"`
		PublicInterestScore float64 `json:"public_interest_score"`

		CommunityData struct {
			FacebookLikes            int `json:"facebook_likes"`
			TwitterFollowers         int `json:"twitter_followers"`
			RedditSubscribers        int `json:"reddit_subscribers"`
			TelegramChannelUserCount int `json:"telegram_channel_user_count"`
		} `json:"community_data"`

		DeveloperData struct {
			Forks                   int `json:"forks"`
			Stars                   int `json:"stars"`
			Subscribers             int `json:"subscribers"`
			TotalIssues             int `json:"total_issues"`
			ClosedIssues            int `json:"closed_issues"`
			PullRequestsMerged      int `json:"pull_requests_merged"`
			PullRequestContributors int `json:"pull_request_contributors"`
			CommitCount4Weeks       int `json:"commit_count_4_weeks"`
		} `json:"developer_data"`

		PublicInterestStats struct {
			AlexaRank   int `json:"alexa_rank"`
			BingMatches int `json:"bing_matches"`
		} `json:"public_interest_stats"`

		LastUpdated time.Time `json:"last_updated"`
	}
)

func (cd *cgkCoinDetail) ToCoinDetail() *AssetDetail {
	r := AssetDetail{}
	r.Asset = fintypes.NewCoin(cd.Name, cd.Symbol, fintypes.CoinGecko)
	r.Explorers = cd.Links.BlockchainSite
	r.Explorers = gstring.RemoveSpaces(r.Explorers)
	r.Logo = cd.Image.Large
	r.Websites = cd.Links.Homepage
	r.Websites = gstring.RemoveSpaces(r.Websites)
	r.SocialMedia = append(r.SocialMedia, cd.Links.AnnouncementUrl...)
	r.SocialMedia = append(r.SocialMedia, cd.Links.ChatUrl...)
	r.SocialMedia = append(r.SocialMedia, cd.Links.ReposUrl.Github...)
	r.SocialMedia = append(r.SocialMedia, cd.Links.ReposUrl.Bitbucket...)
	r.SocialMedia = append(r.SocialMedia, cd.Links.SubRedditUrl)
	r.SocialMedia = append(r.SocialMedia, cd.Links.OfficialForumUrl...)
	r.SocialMedia = gstring.RemoveDuplicate(r.SocialMedia)
	r.SocialMedia = gstring.RemoveSpaces(r.SocialMedia)
	return &r
}

func (c *CgkClient) GetAssets() ([]fintypes.Asset, error) {
	b, err := ghttp.GetBytes("https://api.coingecko.com/api/v3/coins/list", c.proxy, time.Minute*2)
	if err != nil {
		return nil, err
	}
	var list []cgkCoinName
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	var res []fintypes.Asset
	for i := range list {
		res = append(res, fintypes.NewCoin(list[i].Name, list[i].Symbol, fintypes.CoinGecko))
	}
	return res, nil
}

func (c *CgkClient) GetDetail(coin fintypes.Asset) (*AssetDetail, error) {
	uri := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?tickers=false&market_data=false&community_data=true&developer_data=true&sparkline=false", strings.ToLower(coin.Name()))
	b, err := ghttp.GetBytes(uri, c.proxy, time.Minute*2)
	if err != nil {
		return nil, err
	}
	tmp := cgkCoinDetail{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, errors.Wrap(err, "Unmarshal")
	}
	if tmp.Error != "" {
		return nil, errors.Errorf("%s:%s", coin.Name(), tmp.Error)
	}
	detail := tmp.ToCoinDetail()

	var sps []struct {
		Error             string  `json:"error"`
		CirculatingSupply float64 `json:"circulating_supply"`
		TotalSupply       float64 `json:"total_supply"`
	}
	uri = fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=1&page=1&sparkline=false", coin.Name())
	b, err = ghttp.GetBytes(uri, c.proxy, time.Minute*2)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &sps); err != nil {
		return nil, err
	}
	if len(sps) == 1 {
		if sps[0].Error != "" {
			return nil, errors.Errorf("%s:%s", coin.Name(), tmp.Error)
		}
		detail.TotalSupply = sps[0].TotalSupply
		detail.CirculatingSupply = sps[0].CirculatingSupply
	}

	if detail.Github() == "" && len(detail.Websites) > 0 {
		html, err := ghttp.GetString(detail.Websites[0], c.proxy, time.Minute)
		if err == nil {
			doc, err := ghtml.NewDocFromHtmlSrc(&html)
			if err == nil {
				doc.Find("a").Each(func(i int, selection *goquery.Selection) {
					href, exist := selection.Attr("href")
					if exist {
						if strings.Contains(href, "github.com/") {
							if gstring.CountByValue(detail.SocialMedia, href) == 0 {
								detail.SocialMedia = append(detail.SocialMedia, href)
							}
						}
					}
				})
			}
		}
	}

	return detail, nil
}
