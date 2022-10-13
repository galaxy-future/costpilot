package cloud

type Provider string

const Undefined = "undefined"
const (
	AlibabaCloud Provider = "AlibabaCloud"
	HuaweiCloud           = "HuaweiCloud"
	TencentCloud          = "TencentCloud"
	BaiduCloud            = "BaiduCloud"
	AWSCloud              = "AWSCloud"
)

func (p Provider) String() string {
	switch p {
	case AlibabaCloud, HuaweiCloud, TencentCloud, BaiduCloud, AWSCloud:
		return string(p)
	}
	return Undefined
}

func (p Provider) StringCN() string {
	switch p {
	case AlibabaCloud:
		return "阿里云"
	case HuaweiCloud:
		return "华为"
	case TencentCloud:
		return "腾讯"
	case BaiduCloud:
		return "百度"
	case AWSCloud:
		return "AWS"
	}
	return Undefined
}

type SubscriptionType string

const (
	PrePaid  SubscriptionType = "PrePaid"
	PostPaid SubscriptionType = "PostPaid"
)

func (s SubscriptionType) String() string {
	switch s {
	case PrePaid:
		return string(s)
	case PostPaid:
		return string(s)
	}
	return Undefined
}

func (s SubscriptionType) StringCN() string {
	switch s {
	case PrePaid:
		return "包年包月"
	case PostPaid:
		return "按量付费"
	}
	return Undefined
}
