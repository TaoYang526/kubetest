package common

import (
    "fmt"
    "time"
)

func ConvertToTime(beginTime time.Time, endTime time.Time, timeIf interface{}) time.Time {
    if time,ok := timeIf.(time.Time); ok {
        if time.Before(beginTime) {
            panic(fmt.Errorf("Convert failed: time(%v) is previous to the begin time(%v)\n", time, beginTime))
        } else if time.After(endTime) {
            panic(fmt.Errorf("Convert failed: time(%v) is after the end time(%v)\n", time, endTime))
        }
        return time
    } else {
        panic(fmt.Errorf("Convert failed: type of data is not time.Time, data: %v ", timeIf))
    }
}

func ConvertToString(stringIf interface{}) string {
    if str, ok := stringIf.(string); ok {
        return str
    } else {
        panic(fmt.Errorf("Convert failed: type of data is not string, data: %v ", str))
    }
}

func ConvertToInt64(int64If interface{}) int64 {
    if int64var, ok := int64If.(int64); ok {
        return int64var
    } else {
        panic(fmt.Errorf("Convert failed: type of data is not int64, data: %v ", int64If))
    }
}

func ConvertToInt(intIf interface{}) int {
    if intvar, ok := intIf.(int); ok {
        return intvar
    } else {
        panic(fmt.Errorf("Convert failed: type of data is not int, data: %v ", intIf))
    }
}