package util

type CorrelationIdKey string

func (c CorrelationIdKey) String() string {
	return string(c)
}
