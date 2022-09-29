package number

import (
	"strconv"
	"strings"
)

// FormatPrice 价格格式化，最多两位有效数字，去掉所有小数后最后的0
func FormatPrice(price float64) string {
	result := strconv.FormatFloat(price, 'f', 2, 64)
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0")
		return strings.TrimRight(result, ".")
	}
	return result
}
