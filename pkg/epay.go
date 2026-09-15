package pkg

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func Sign(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	b.WriteString(key)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func Verify(params map[string]string, key, sign string) bool {
	if sign == "" || key == "" {
		return false
	}
	return strings.EqualFold(Sign(params, key), sign)
}

func NewOrderID() string {
	var n [4]byte
	_, _ = rand.Read(n[:])
	return fmt.Sprintf("ORD%s%s", time.Now().UTC().Format("20060102150405"), hex.EncodeToString(n[:]))
}

func NewCardKey() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("R%X", time.Now().UnixNano())
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return fmt.Sprintf("%s-%s-%s", b[0:4], b[4:8], b[8:12])
}

func SubmitURL() string {
	u := strings.TrimSpace(os.Getenv("GOPAY_SUBMIT_URL"))
	if u == "" {
		u = "https://pay.hunyuantaiji.shop/xpay/epay/submit.php"
	}
	return u
}

func BuildPayURL(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	base := SubmitURL()
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + q.Encode()
}

func PublicBase(rHost, proto string) string {
	if proto == "" {
		proto = "https"
	}
	host := strings.TrimSpace(rHost)
	if i := strings.Index(host, ","); i >= 0 {
		host = strings.TrimSpace(host[:i])
	}
	return proto + "://" + host
}
