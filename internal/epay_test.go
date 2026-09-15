package internal

import "testing"

func TestSignMatchesEpayDoc(t *testing.T) {
	params := map[string]string{
		"money":        "1",
		"name":         "【A0125】",
		"notify_url":   "https://www.zhifux.com/success.txt",
		"out_trade_no": "38329329329121",
		"pid":          "442130427265040384",
		"return_url":   "https://www.baidu.com",
		"type":         "alipay",
	}
	got := Sign(params, "072c320169c1c4043277c8746c26050f")
	if got != "134d9c7103b5534ea72935542c9e4a31" {
		t.Fatalf("sign=%s", got)
	}
}
