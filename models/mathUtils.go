package models

import "sort"

func getCeilingDivisible(num int, divisibleBy int) int {
	if num%divisibleBy == 0 {
		return num
	}
	return ((num / divisibleBy) + 1) * divisibleBy
}

func getAllDivisibleNumbers(dataNodes int, azs int) (li []int) {
	if dataNodes <= 0 || azs <= 0 {
		return
	}
	for i := 1; i <= dataNodes; i++ {
		if dataNodes%i == 0 && i >= azs && i%azs == 0 {
			li = append(li, i)
			if azs == 1 {
				continue
			}
			for j := 1; j <= azs; j++ {
				li = append(li, i*j, (i+dataNodes)*j)
			}
		}
	}
	li = removeDuplicateValues(li)
	sort.Ints(li[:])
	return li
}

func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}

func removeDuplicateValues(intSlice []int) []int {
	keys := make(map[int]bool)
	var list []int

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
