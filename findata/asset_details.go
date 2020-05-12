package findata

import (
	"encoding/json"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/net/gaddr"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"strings"
)

var (
	// Supply
	TagDynamicSupply = "DynamicSupply"

	// Stable Coin
	TagFiatAnchored = "FiatAnchored"

	// Main net
	TagNative = "native"
	TagERC20  = "erc20" // Ethereum ERC20 token
	TagWRC20  = "wrc20" // WanChain WRC20 token
	TagDapp   = "dapp"

	// Industry
	TagCrossChain    = "CrossChain"
	TagPublicChain   = "PublicChain"
	TagSideChain     = "SideChain"
	TagCloudCompute  = "CloudCompute"
	TagOS            = "OS"
	TagIoT           = "IoT"
	TagIM            = "IM"
	TagBigData       = "BigData"
	TagAI            = "AI"
	TagCopyright     = "Copyright"
	TagStorage       = "Storage"
	TagAnonymous     = "Anonymous"
	TagContent       = "Content"
	TagExchange      = "Exchange"
	TagGame          = "Game"
	TagSmartContract = "SmartContract"
	TagNotary        = "Notary" // 公证防伪

	ErrNotIncludedYet = gerror.New("Not included yet")
	ErrTryAgainLater  = gerror.New("Try again later, wait network download done")
	ErrNameNotFound   = gerror.New("UniqueName not found") // 由于平台上某些币的登记名字差别、有没有登记的差别，导致找不到对应页面
	ErrDDosProtection = gerror.New("DDos Protection")      // http access too frequently
	ErrNoHistoryData  = gerror.New("No history data")      // 在CMC上某些币有概况，但是历史数据没有登记，比如fcoin-token
	ErrNoIcoInfo      = gerror.New("No ICO info")          // feixiaohao上有些币的ICO信息未收录
)

type (
	Ico struct {
		AvgPrice float64          `json:"AvgPrice"` // ICO发行价，如果有多个价格则选择均价
		Amount   float64          `json:"Amount"`   // 募资总额
		Date     *gtime.DateRange `json:"Date"`
	}

	AssetDetail struct {
		Asset             fintypes.Asset `json:"Asset" bson:"_id"`
		MaxSupply         float64        `json:"MaxSupply,omitempty" bson:"MaxSupply,omitempty"`
		TotalSupply       float64        `json:"TotalSupply,omitempty" bson:"TotalSupply,omitempty"`
		CirculatingSupply float64        `json:"CirculatingSupply,omitempty" bson:"CirculatingSupply,omitempty"`
		ContractAddress   string         `json:"ContractAddress,omitempty" bson:"ContractAddress,omitempty"`

		// not useful for now
		Tags            []string            `json:"Tags,omitempty" bson:"Tags,omitempty"`
		ListedExchanges map[string][]string `json:"ListedExchanges,omitempty" bson:"ListedExchanges,omitempty"`

		// links
		Logo        string   `json:"Logo,omitempty" bson:"Logo,omitempty"`
		Websites    []string `json:"Websites,omitempty" bson:"Websites,omitempty"`
		Explorers   []string `json:"Explorers,omitempty" bson:"Explorers,omitempty"`
		TechDocs    []string `json:"TechDocs,omitempty" bson:"TechDocs,omitempty"`
		SocialMedia []string `json:"SocialMedia,omitempty" bson:"SocialMedia,omitempty"`
		Exchanges   []string `json:"Exchanges,omitempty" bson:"Exchanges,omitempty"`
	}
)

func (c *AssetDetail) Github() string {
	for _, v := range c.SocialMedia {
		if strings.Contains(strings.ToLower(v), "github.com") {
			return v
		}
	}
	return ""
}

func (c *AssetDetail) Website() string {
	if len(c.Websites) > 0 {
		return c.Websites[0]
	} else {
		return ""
	}
}

func (c *AssetDetail) TechDoc() string {
	if len(c.TechDocs) > 0 {
		return c.TechDocs[0]
	} else {
		return ""
	}
}

func (c *AssetDetail) AppendSocialMedia(newurls ...string) {
	c.SocialMedia = append(c.SocialMedia, newurls...)
	c.SocialMedia = gaddr.RemoveDuplicateUrl(c.SocialMedia)
}

func (c *AssetDetail) GetSocialMediaContains(contain string) []string {
	var r []string

	for _, v := range c.SocialMedia {
		if strings.Contains(strings.ToLower(v), strings.ToLower(contain)) {
			r = append(r, v)
		}
	}
	return r
}

func (c *AssetDetail) String() string {
	buf, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (c *AssetDetail) Update(newinfo AssetDetail) {
	c.TotalSupply = newinfo.TotalSupply
	c.CirculatingSupply = newinfo.CirculatingSupply
	c.ContractAddress = newinfo.ContractAddress
	//c.IcoAvgPrice = newinfo.IcoAvgPrice
	//c.IcoAmount = newinfo.IcoAmount
	//c.IcoDate = newinfo.IcoDate
	//c.Tags = newinfo.Tags
	c.Websites = newinfo.Websites
	c.Logo = newinfo.Logo
	c.Explorers = newinfo.Explorers
	c.TechDocs = newinfo.TechDocs
	c.SocialMedia = newinfo.SocialMedia
}

/*
// 将本地人工配置的CryptoInfo，Merge到网上获取的CryptoInfo
func (c *AssetDetail) Merge(extended AssetDetail) {
	if extended.TotalSupply > 0 {
		c.TotalSupply = extended.TotalSupply
	}
	if extended.CirculatingSupply > 0 {
		c.CirculatingSupply = extended.CirculatingSupply
	}
	_, err := eth.ParseAddress(extended.ContractAddress)
	if err == nil {
		c.ContractAddress = extended.ContractAddress
	}
	/*if extended.IcoAvgPrice.BiggerThanZero() && extended.IcoAmount.BiggerThanZero() {
		c.IcoAvgPrice = extended.IcoAvgPrice
		c.IcoAmount = extended.IcoAmount
	}
	if !extended.IcoDate.IsZero() {
		c.IcoDate = extended.IcoDate
	}
	if len(extended.Tags) > 0 {
		c.Tags = append(c.Tags, extended.Tags...)
		c.Tags = xstring.ToLower(c.Tags)
		c.Tags = xstring.RemoveDuplicate(c.Tags)
	}*/
/*
	// Links
	if len(extended.Websites) > 0 {
		c.Websites = append(c.Websites, extended.Websites...)
		c.Websites = addr.RemoveDuplicateUrl(c.Websites)
	}
	if len(extended.TechDocs) > 0 {
		c.TechDocs = append(c.TechDocs, extended.TechDocs...)
		c.TechDocs = addr.RemoveDuplicateUrl(c.TechDocs)
	}
	if len(c.Logo) == 0 && len(extended.Logo) > 0 {
		c.Logo = extended.Logo
	}
	if len(extended.Explorers) > 0 {
		c.Explorers = append(c.Explorers, extended.Explorers...)
		c.Explorers = addr.RemoveDuplicateUrl(c.Explorers)
	}
	if len(extended.SocialMedia) > 0 {
		c.SocialMedia = append(c.SocialMedia, extended.SocialMedia...)
		c.SocialMedia = addr.RemoveDuplicateUrl(c.SocialMedia)
	}
}*/
