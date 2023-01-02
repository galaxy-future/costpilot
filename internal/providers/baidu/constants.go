package baidu

type regionType string

var (
	// https://cloud.baidu.com/doc/BCC/s/0jwvyo603
	_endPointMap = map[string]string{
		"bj":  ".bj.baidubce.com",
		"gz":  ".gz.baidubce.com",
		"su":  ".su.baidubce.com",
		"hkg": ".hkg.baidubce.com",
		"fwh": ".fwh.baidubce.com",
		"bd":  ".bd.baidubce.com",
		"sin": ".sin.baidubce.com",
		"fsh": ".fsh.baidubce.com",
	}

	_regionNameMap = map[string]string{
		"bj":  "北京",
		"gz":  "广州",
		"su":  "苏州",
		"hkg": "香港",
		"fwh": "武汉",
		"bd":  "保定",
		"sin": "新加坡",
		"fsh": "上海",
	}
)
