package aws

var _regionLocalName = map[string]string{
	"cn-north-1":     "中国 (北京)",
	"cn-northwest-1": "中国 (宁夏)",
	"us-east-1":      "美国东部 (弗吉尼亚北部)",
	"us-east-2":      "美国东部 (俄亥俄州)",
	"us-west-1":      "美国西部 (加利福尼亚北部)",
	"us-west-2":      "美国西部 (俄勒冈州)",
	"af-south-1":     "非洲 (开普敦)",
	"ap-east-1":      "亚太地区 (香港)",
	"ap-southeast-3": "亚太地区 (雅加达)",
	"ap-south-1":     "亚太地区 (孟买)",
	"ap-northeast-3": "亚太地区 (大阪)",
	"ap-northeast-2": "亚太地区 (首尔)",
	"ap-southeast-1": "亚太地区 (新加坡)",
	"ap-southeast-2": "亚太地区 (悉尼)",
	"ap-northeast-1": "亚太地区 (东京)",
	"ca-central-1":   "加拿大 (中部)",
	"eu-central-1":   "欧洲 (法兰克福)",
	"eu-west-1":      "欧洲 (爱尔兰)",
	"eu-west-2":      "欧洲 (伦敦)",
	"eu-south-1":     "欧洲 (米兰)",
	"eu-west-3":      "欧洲 (巴黎)",
	"eu-north-1":     "欧洲 (斯德哥尔摩)",
	"me-south-1":     "中东 (巴林)",
	"sa-east-1":      "南美洲 (圣保罗)",
}

const (
	CPUUtilization    string = "CPUUtilization"
	MemoryUtilization string = "mem_used_percent"
	Namespace_Cpu     string = "AWS/EC2"
	Namespace_Mem     string = "CWAgent"
	InstanceId        string = "InstanceId"
)
