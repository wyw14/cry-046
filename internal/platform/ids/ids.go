package ids

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Sequence struct{ counter uint64 }

func (s *Sequence) NewID(prefix string) string {
	n := atomic.AddUint64(&s.counter, 1)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), n)
}
