package tools

func CurrencyUnit(name string) (result string) {
	switch name {
	case "USD":
		result = "$"
	case "CNY":
		result = "¥"
	}
	return
}
