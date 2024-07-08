package session

import (
	"atmo/util/str"
	"hash/adler32"
	"hash/crc32"
	"hash/maphash"
)

var (
	hasherMh    = maphash.Hash{}
	hasherCrc   = crc32.NewIEEE()
	hasherAdler = adler32.New()
)

func hashAdler(src string) string {
	hasherAdler.Reset()
	hasherAdler.Write([]byte(src))
	return str.FromU64(uint64(hasherAdler.Sum32()), 36)
}

func hashMh(src string) string {
	hasherMh.Reset()
	hasherMh.Write([]byte(src))
	return str.FromU64(hasherMh.Sum64(), 36)
}

func hashCrc(src string) string {
	hasherCrc.Reset()
	hasherCrc.Write([]byte(src))
	return str.FromU64(uint64(hasherCrc.Sum32()), 36)
}

func contentHash(src string) string {
	return hashAdler(src) + hashMh(src) + hashCrc(src)
}

func contentHashEq(src1 string, src2 string) string {
	adler := hashAdler(src1)
	if hashAdler(src2) != adler {
		return ""
	}
	mh := hashMh(src1)
	if hashMh(src2) != mh {
		return ""
	}
	crc := hashCrc(src1)
	if hashCrc(src2) != crc {
		return ""
	}
	return adler + mh + crc
}
