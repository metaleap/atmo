package kv

type Any = map[string]any
type Of[T any] map[string]T

func Keys[K comparable, V any](m map[K]V) (ret []K) {
	ret = make([]K, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return
}

func Values[K comparable, V any](m map[K]V) (ret []V) {
	ret = make([]V, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return
}

func Fill[K comparable, V any](dst map[K]V, from map[K]V) map[K]V {
	for k, v := range from {
		dst[k] = v
	}
	return dst
}

func FromKeys[TKey comparable, TVal any](keys []TKey, value func(TKey) TVal) map[TKey]TVal {
	ret := make(map[TKey]TVal, len(keys))
	for _, key := range keys {
		ret[key] = value(key)
	}
	return ret
}

func FromValues[TKey comparable, TVal any](values []TVal, key func(TVal) TKey) map[TKey]TVal {
	ret := make(map[TKey]TVal, len(values))
	for _, val := range values {
		ret[key(val)] = val
	}
	return ret
}

func Eq[K comparable, V any](opl map[K]V, opr map[K]V, eq func(opl V, opr V) bool) bool {
	if len(opl) != len(opr) {
		return false
	}
	for k, v := range opl {
		if found, exists := opr[k]; (!exists) || !eq(v, found) {
			return false
		}
	}
	return true
}

func To[TKey comparable, TValIn any, TValOut any](m map[TKey]TValIn, f func(TValIn) TValOut) map[TKey]TValOut {
	ret := make(map[TKey]TValOut, len(m))
	for k, v := range m {
		ret[k] = f(v)
	}
	return ret
}
