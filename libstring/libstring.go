package libstring

import (
	"encoding/json"
	"math"
	"strings"
	"unicode"
)

// JSONEncode - Encode JSON without escape HTML
func JSONEncode(data interface{}) string {
	bt, _ := json.Marshal(data)
	return string(bt)
}

// Ucfirst - Upper case first character
func Ucfirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToUpper(v))
		return u + str[len(u):]
	}
	return ""
}

// KeyToLabel - Format underscore key string to value string
func KeyToLabel(input string) string {
	output := strings.Replace(input, "_", " ", -1)
	return strings.Title(output)
}

// LabelToKey - Format string by replace space to underscore
func LabelToKey(input string) string {
	output := strings.Replace(input, " ", "_", -1)
	return strings.ToLower(output)
}

// TitleToNumber - Convert A to 0, AA to 26
func TitleToNumber(s string) int {
	weight := 0.0
	sum := 0
	for i := len(s) - 1; i >= 0; i-- {
		ch := int(s[i])
		if int(s[i]) >= int('a') && int(s[i]) <= int('z') {
			ch = int(s[i]) - 32
		}
		sum = sum + (ch-int('A')+1)*int(math.Pow(26, weight))
		weight++
	}
	return sum - 1
}

// ToURL - Encode string to URL
func ToURL(s string) string {
	s = strings.Trim(s, " ")
	s = strings.ReplaceAll(s, " ", "%20")
	return s
}
