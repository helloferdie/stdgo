package libslice

import (
	"fmt"
	"strconv"
	"strings"
)

// Contains - Check slice contains value of string
func Contains(a string, list []string) (int, bool) {
	for k, b := range list {
		if b == a {
			return k, true
		}
	}
	return -1, false
}

// ContainsInt - Check slice contains value of int
func ContainsInt(a int, list []int) (int, bool) {
	for k, b := range list {
		if b == a {
			return k, true
		}
	}
	return -1, false
}

// ContainsInt64 - Check slice contains value of int64
func ContainsInt64(a int64, list []int64) (int, bool) {
	for k, b := range list {
		if b == a {
			return k, true
		}
	}
	return -1, false
}

// Unique - Return unique slice of string
func Unique(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueInt - Return unique slice of int
func UniqueInt(s []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueInt64 - Return unique slice of int64
func UniqueInt64(s []int64) []int64 {
	keys := make(map[int64]bool)
	list := []int64{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// SliceIntToString - Convert slice int to string
func SliceIntToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

// SliceStringToString - Convert slice string to string
func SliceStringToString(a []string, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

// SliceInterfaceToSliceString - Convert slice interface to slice string
func SliceInterfaceToSliceString(a []interface{}) []string {
	output := []string{}
	for _, v := range a {
		output = append(output, v.(string))
	}
	return output
}

// SliceInterfaceToSliceInt - Convert slice interface to slice int
func SliceInterfaceToSliceInt(a []interface{}) []int {
	output := []int{}
	for _, v := range a {
		str, _ := strconv.Atoi(v.(string))
		output = append(output, str)
	}
	return output
}

// Insert - Insert at random index
func Insert(a []interface{}, index int, value interface{}) []interface{} {
	if len(a) <= index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

// Remove - Remove string from slice string
func Remove(src []string, txt string, keepOrder bool) []string {
	k, exist := Contains(txt, src)
	if exist {
		if keepOrder {
			return append(src[:k], src[k+1:]...)
		}
		src[k] = src[len(src)-1]
		return src[:len(src)-1]
	}
	return src
}
