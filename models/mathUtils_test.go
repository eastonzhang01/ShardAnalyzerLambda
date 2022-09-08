package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getPossibleNumberWithInvalidParam(t *testing.T) {
	li := getAllDivisibleNumbers(2, 0)
	assert.True(t, li == nil, "list should be empty")

	li = getAllDivisibleNumbers(0, -0)
	assert.True(t, li == nil, "list should be empty")
}

func Test_getPossibleNumberWithValidParams(t *testing.T) {
	li := getAllDivisibleNumbers(4, 2)
	assert.ElementsMatch(t, li, []int{2, 4, 6, 8, 12, 16})

	li = getAllDivisibleNumbers(6, 2)
	assert.ElementsMatch(t, li, []int{2, 4, 6, 8, 12, 16, 24})

	li = getAllDivisibleNumbers(6, 1)
	assert.ElementsMatch(t, li, []int{1, 2, 3, 6})

	li = getAllDivisibleNumbers(10, 1)
	assert.ElementsMatch(t, li, []int{1, 2, 5, 10})

	li = getAllDivisibleNumbers(15, 3)
	assert.ElementsMatch(t, li, []int{3, 6, 9, 15, 18, 30, 36, 45, 54, 60, 90})
}
