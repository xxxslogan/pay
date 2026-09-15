package handler

import (
	"encoding/json"
	"fmt"
	"html"
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

type activateReq struct {
	CardKey  string `json:"cardKey"`
	DeviceID string `json:"deviceId"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch route(r) {
	case "getPayUrl":
		getPayUrl(w, r)
	case "notify":
		notify(w, r)
	case "activate":
		activate(w, r)
	case "paid":
		paid(w, r)
	default:
		if pkg.CORS(w, r) {
			return
		}
		pkg.JSON(w, 200, map[string]any{"code": 404, "msg": "unknown route"})
	}
}

func route(r *http.Request) string {
	p := strings.ToLower(strings.TrimSuffix(r.URL.Path, "/"))
	switch {
	case strings.HasSuffix(p, "/getpayurl"):
		return "getPayUrl"
	case strings.HasSuffix(p, "/notify"):
		return "notify"
	case strings.HasSuffix(p, "/activate"):
		return "activate"
	case strings.HasSuffix(p, "/paid"):
		return "paid"
	default:
		return ""
	}
}

func getPayUrl(w http.ResponseWriter, r *http.Request) {
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
	money := pkg.PackageMoney(req.PackageID)
	params := map[string]string{
		"pid":          pid,
		"type":         "wxpay",
		"out_trade_no": orderID,
		"notify_url":   base + "/api/notify",
		"return_url":   base + "/api/paid",
		"name":         pkg.PackageName(req.PackageID),
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

func notify(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	params := map[string]string{}
	for k, vs := range r.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	secret := strings.TrimSpace(os.Getenv("GOPAY_SECRET"))
	sign := params["sign"]
	status := params["trade_status"]
	orderID := params["out_trade_no"]
	money := params["money"]
	if orderID == "" || !pkg.Verify(params, secret, sign) || status != "TRADE_SUCCESS" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fail"))
		return
	}
	var order pkg.Order
	if err := pkg.GetJSON("order:"+orderID, &order); err != nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fail"))
		return
	}
	if order.Money != "" && money != "" && order.Money != money {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fail"))
		return
	}
	if order.Status == "paid" && order.CardKey != "" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
		return
	}
	cardKey := pkg.NewCardKey()
	card := pkg.Card{
		CardKey:    cardKey,
		OrderID:    orderID,
		BindDevice: "",
		ChangeLeft: 1,
		Created:    time.Now().Unix(),
	}
	order.Status = "paid"
	order.CardKey = cardKey
	if err := pkg.SetJSON("card:"+cardKey, card); err != nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fail"))
		return
	}
	if err := pkg.SetJSON("order:"+orderID, order); err != nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fail"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func activate(w http.ResponseWriter, r *http.Request) {
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
	var req activateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSON(w, 200, map[string]any{"code": 400, "msg": "invalid json"})
		return
	}
	req.CardKey = strings.TrimSpace(req.CardKey)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.CardKey == "" || req.DeviceID == "" {
		pkg.JSON(w, 200, map[string]any{"code": 400, "msg": "missing cardKey or deviceId"})
		return
	}
	var card pkg.Card
	if err := pkg.GetJSON("card:"+req.CardKey, &card); err != nil {
		pkg.JSON(w, 200, map[string]any{"code": 1001, "msg": "卡密不存在"})
		return
	}
	if card.BindDevice == "" {
		card.BindDevice = req.DeviceID
		if err := pkg.SetJSON("card:"+req.CardKey, card); err != nil {
			pkg.JSON(w, 200, map[string]any{"code": 500, "msg": "bind fail"})
			return
		}
		pkg.JSON(w, 200, map[string]any{
			"code": 0, "msg": "激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
		})
		return
	}
	if card.BindDevice == req.DeviceID {
		pkg.JSON(w, 200, map[string]any{
			"code": 0, "msg": "激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
		})
		return
	}
	if card.ChangeLeft <= 0 {
		pkg.JSON(w, 200, map[string]any{"code": 1002, "msg": "换机次数耗尽"})
		return
	}
	card.BindDevice = req.DeviceID
	card.ChangeLeft--
	if err := pkg.SetJSON("card:"+req.CardKey, card); err != nil {
		pkg.JSON(w, 200, map[string]any{"code": 500, "msg": "rebind fail"})
		return
	}
	pkg.JSON(w, 200, map[string]any{
		"code": 0, "msg": "换机激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
	})
}

func paid(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("out_trade_no")
	if orderID == "" {
		orderID = r.URL.Query().Get("orderId")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if orderID == "" {
		_, _ = w.Write([]byte(pageHTML("未找到订单号。请回到 RimixFac，支付完成后本页会显示卡密。")))
		return
	}
	var order pkg.Order
	if err := pkg.GetJSON("order:"+orderID, &order); err != nil {
		_, _ = w.Write([]byte(pageHTML("订单尚未同步，请稍候刷新。")))
		return
	}
	if order.Status != "paid" || order.CardKey == "" {
		_, _ = fmt.Fprintf(w, pageHTML(`支付确认中，请稍候…<meta http-equiv="refresh" content="2">订单号：%s`), html.EscapeString(orderID))
		return
	}
	_, _ = fmt.Fprintf(w, pageHTML(`支付成功。请复制卡密，回到 RimixFac 粘贴激活。<div style="margin:24px 0;padding:16px;border:1px dashed #c9a227;font-size:22px;letter-spacing:2px">%s</div>订单号：%s`), html.EscapeString(order.CardKey), html.EscapeString(orderID))
}

func pageHTML(inner string) string {
	return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>RimixFac 授权</title>
<style>body{font-family:"Microsoft YaHei",sans-serif;background:#1a140c;color:#f3e6d0;display:flex;min-height:100vh;align-items:center;justify-content:center}
.card{max-width:520px;background:#241c14;padding:32px;border:1px solid #3d3224;line-height:1.6}</style></head>
<body><div class="card"><h2>RimixFac 永久授权</h2>` + inner + `</div></body></html>`
}
