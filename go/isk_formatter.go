package main

import (
	"fmt"
	"log"
	"math"
)

var suffix = map[int]string{
	1: "K",
	2: "M",
	3: "B",
	4: "T",
}

func Format(value float64) string {
	if value < 1000 {
		return fmt.Sprintf("%.2f", value)
	}

	count := 0
	tmpValue := value
	for 1000 < tmpValue {
		count += 1
		tmpValue /= 1000
	}

	log.Println(count)

	return fmt.Sprintf("%.2f%s", value/math.Pow(1000, float64(count)), suffix[count])
}
