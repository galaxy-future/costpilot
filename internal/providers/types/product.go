package types

type PipCode string

func (p PipCode) String() string {
	return string(p)
}

// Notice: 暂不使用 PipCode 统一归类
const (
	ECS     PipCode = "ecs"
	EIP     PipCode = "eip"
	GWS     PipCode = "gws"
	KVSTORE PipCode = "kvstore"
	DISK    PipCode = "disk"
	NAS     PipCode = "nas"
	NAT     PipCode = "nat"
	S3      PipCode = "s3"
	SLB     PipCode = "slb"
	CBN     PipCode = "cbn"
)

const Undefined = "Undefined"

func PidCode2Name(pipCode PipCode) string {
	switch pipCode {
	case ECS:
		return "云服务器"
	case EIP:
		return "弹性公网IP"
	case GWS:
		return "云桌面"
	case KVSTORE:
		return "云数据库Redis"
	case DISK:
		return "块存储"
	case NAS:
		return "文件存储NAS"
	case NAT:
		return "NAT网关"
	case S3:
		return "对象存储"
	case SLB:
		return "负载均衡"
	case CBN:
		return "云企业网"
	}
	return Undefined
}
