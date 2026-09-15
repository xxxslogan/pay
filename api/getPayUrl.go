package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xxxslogan/pay/pkg"
)

type payReq struct {
	PackageID string `json:"packageId"`
	DeviceID  string `json:"deviceId"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if pkg.CORS(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		pkg.JSON(w, 200, map[string]any{"code": 405, "msg": "method not allowed"})
		return
	}
	if !pkg.CheckAPIToken(r) {
		pkg.JSON(w, 200, map[string]any{"code": 401, "msg": "unauthorized"})
		return
	}
	var req payReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSON(w, 200, map[string]any{"code": 400, "msg": "json parse error"})
		return
	}
	req.PackageID = strings.TrimSpace(req.PackageID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" || !pkg.PackageOK(req.PackageID) {
		pkg.JSON(w, 200, map[string]any{"code": 400, "msg": "invalid packageId or deviceId"})
		return
	}
	pid := strings.TrimSpace(os.Getenv("GOPAY_PID"))
	secret := strings.TrimSpace(os.Getenv("GOPAY_SECRET"))
	if pid == "" || secret == "" {
		pkg.JSON(w, 200, map[string]any{"code": 500, "msg": "pay config missing"})
		return
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "https"
	}
	base := pkg.PublicBase(host, proto)
	orderID := pkg.NewOrderID()
	money := pkg.PackageMoney()
	params := map[string]string{
		"pid":          pid,
		"type":         "wxpay",
		"out_trade_no": orderID,
		"notify_url":   base + "/api/notify",
		"return_url":   base + "/api/paid",
		"name":         pkg.PackageName(),
		"money":        money,
		"param":        req.PackageID,
	}
	params["sign"] = pkg.Sign(params, secret)
	params["sign_type"] = "MD5"
	order := pkg.Order{
		OrderID:   orderID,
		DeviceID:  req.DeviceID,
		PackageID: req.PackageID,
		Money:     money,
		Status:    "pending",
		Created:   time.Now().Unix(),
	}
	if err := pkg.SetJSON("order:"+orderID, order); err != nil {
		pkg.JSON(w, 200, map[string]any{"code": 500, "msg": "store order fail"})
		return
	}
	pkg.JSON(w, 200, map[string]any{
		"code":    0,
		"msg":     "ok",
		"payUrl":  pkg.BuildPayURL(params),
		"orderId": orderID,
	})
}
