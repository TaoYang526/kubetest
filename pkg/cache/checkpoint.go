package cache

import (
    "reflect"
    "time"
)

type Checkpoint struct {
    Time         time.Time
    Seconds      int
    MetricValues []int
}

func (cp *Checkpoint) HasSameMetricValues(otherCP *Checkpoint) bool {
    if reflect.DeepEqual(cp.MetricValues, otherCP.MetricValues) {
        return true
    } else {
        return false
    }
}

