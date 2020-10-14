package model

import (
	"fmt"
	"sync"
)

//椅子
type Chair struct {
	chair []int32
	l     sync.Mutex
}

func (c *Chair) String() string {
	return fmt.Sprintf("%v", c.chair)
}

func (c *Chair) OnInit(chairNum int32) {
	for i := int32(0); i < chairNum; i++ {
		c.chair = append(c.chair, i)
	}
}

func (c *Chair) IsFull() bool {
	if len(c.chair) <= 0 {
		return true
	}
	return false
}

func (c *Chair) GetChair() int32 {
	c.l.Lock()
	defer c.l.Unlock()

	if c.IsFull() {
		return -1
	}
	chair := c.chair[0]
	c.chair = c.chair[1:]
	return chair

}

func (c *Chair) AddChair(chair int32) {

	c.l.Lock()
	defer c.l.Unlock()

	c.chair = append(c.chair, chair)

}
