package cluster

import (
	"encoding/json"
	"time"
)

var (
	pingTime = time.Second * 30
	lockTTL  = time.Second * 3
)

const (
	redisClusterKey    = "cluster:%s"
	redisLocksKey      = "cluster:%s:locks:%s"
	redisIncrementProp = "nodeId"
	redisChannelKey    = "channel:%s"
)

type messageType int

const (
	ping          messageType = iota // 0
	nodeJoined                       // 1
	nodeLeave                        // 2
	nodeBroadcast                    // 3
)

type Message struct {
	Type   messageType     `json:"type"`
	NodeID int             `json:"nodeID"`
	Data   json.RawMessage `json:"data"`
}

func (m *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m Message) Int() (res int, err error) {
	err = m.Unpack(&res)
	return
}

func (m Message) String() (res string, err error) {
	err = m.Unpack(&res)
	return
}

func (m Message) Bool() (res bool, err error) {
	err = m.Unpack(&res)
	return
}

func (m Message) Float() (res float64, err error) {
	err = m.Unpack(&res)
	return
}

func (m Message) Unpack(ptr interface{}) (err error) {
	return json.Unmarshal(m.Data, &ptr)
}
