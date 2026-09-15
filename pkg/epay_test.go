package pkg

import "testing"

func TestPackageCatalog(t *testing.T) {
	if !PackageOK("ace-perm-399") || !PackageOK("test-1") || PackageOK("nope") {
		t.Fatal("package allow-list")
	}
	if PackageMoney("test-1") != "1.00" {
		t.Fatalf("test money=%s", PackageMoney("test-1"))
	}
	if PackageMoney("ace-perm-399") != "399.00" {
		t.Fatalf("live money=%s", PackageMoney("ace-perm-399"))
	}
	if PackageName("test-1") == PackageName("ace-perm-399") {
		t.Fatal("test package should use a distinct name")
	}
}

func TestSameMoney(t *testing.T) {
	if !SameMoney("1.00", "1") || !SameMoney("1", "1.0") || !SameMoney("399.00", "399") {
		t.Fatal("decimal money should match")
	}
	if SameMoney("1.00", "1.01") || SameMoney("399.00", "1") {
		t.Fatal("different money must not match")
	}
}

func TestTradePaid(t *testing.T) {
	if !TradePaid("TRADE_SUCCESS") || !TradePaid("success") || TradePaid("TRADE_CLOSED") {
		t.Fatal("trade status")
	}
}

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
