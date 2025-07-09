package mocks

import (
	"strconv"
	"sync"

	"github.com/lansfy/gonkex/colorize"
)

const OrderNoValue = -1

type orderChecker struct {
	value int
	mutex sync.Mutex
}

func newOrderChecker() *orderChecker {
	return &orderChecker{
		value: OrderNoValue,
	}
}

func (c *orderChecker) Update(value int) error {
	if value == OrderNoValue {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error
	if c.value > value {
		err = colorize.NewError("the %s of the current request (%s) is less than that of the previous request (%s)",
			colorize.Cyan("order"),
			colorize.Red(strconv.Itoa(value)),
			colorize.Green(strconv.Itoa(c.value)),
		)
	}
	c.value = value
	return err
}

func (c *orderChecker) Reset() {
	c.mutex.Lock()
	c.value = OrderNoValue
	c.mutex.Unlock()
}
