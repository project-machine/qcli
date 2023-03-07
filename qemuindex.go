package qcli

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/yourbasic/bit"
)

type QemuIndex struct {
	bits *bit.Set
}

func NewQemuIndex() *QemuIndex {
	return &QemuIndex{bits: bit.New()}
}

func (q *QemuIndex) Set(index int) error {
	if q.bits.Contains(index) {
		return fmt.Errorf("Index %d already set", index)
	}
	newBits := q.bits.Add(index)
	q.bits = newBits
	return nil
}

func (q *QemuIndex) Next() int {
	if q.bits.Empty() {
		q.Set(0)
		return 0
	}
	// one bit is set
	next := q.bits.Max() + 1
	q.Set(next)
	return next
}

func (q *QemuIndex) String() string {
	return fmt.Sprintf("%s", q.bits)
}

type QemuTypeIndex map[string]*QemuIndex

func NewQemuTypeIndex() *QemuTypeIndex {
	qti := make(QemuTypeIndex)
	return &qti
}

func (qti QemuTypeIndex) Next(qtype string) int {
	if qtype == "" {
		log.Errorf("Invalid empty qtype parameter")
		return -1
	}
	if _, ok := qti[qtype]; !ok {
		qti[qtype] = NewQemuIndex()
	}
	return qti[qtype].Next()
}

func (qti QemuTypeIndex) Set(qtype string, index int) error {
	if qtype == "" {
		return fmt.Errorf("Invalid empty qtype parameter")
	}
	if _, ok := qti[qtype]; !ok {
		qti[qtype] = NewQemuIndex()
	}
	return qti[qtype].Set(index)
}

func (qti QemuTypeIndex) NextBootIndex() int {
	return qti.Next("bootindex")
}

func (qti QemuTypeIndex) SetBootIndex(index int) error {
	return qti.Set("bootindex", index)
}

func (qti QemuTypeIndex) NextDriveIndex() int {
	return qti.Next("drive")
}

func (qti QemuTypeIndex) SetDriveIndex(index int) error {
	return qti.Set("drive", index)
}

func (qti QemuTypeIndex) NextNetIndex() int {
	return qti.Next("net")
}

func (qti QemuTypeIndex) SetNetIndex(index int) error {
	return qti.Set("net", index)
}
