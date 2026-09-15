package handler

import (
	"fmt"
	"html"
	"net/http"

	"github.com/xxxslogan/pay/internal"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("out_trade_no")
	if orderID == "" {
		orderID = r.URL.Query().Get("orderId")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if orderID == "" {
		_, _ = w.Write([]byte(page("未找到订单号。请回到 RimixFac，支付完成后本页会显示卡密。")))
		return
	}
	var order internal.Order
	if err := internal.GetJSON("order:"+orderID, &order); err != nil {
		_, _ = w.Write([]byte(page("订单尚未同步，请稍候刷新。")))
		return
	}
	if order.Status != "paid" || order.CardKey == "" {
		_, _ = fmt.Fprintf(w, page(`支付确认中，请稍候…<meta http-equiv="refresh" content="2">订单号：%s`), html.EscapeString(orderID))
		return
	}
	_, _ = fmt.Fprintf(w, page(`支付成功。请复制卡密，回到 RimixFac 粘贴激活。<div style="margin:24px 0;padding:16px;border:1px dashed #c9a227;font-size:22px;letter-spacing:2px">%s</div>订单号：%s`), html.EscapeString(order.CardKey), html.EscapeString(orderID))
}

func page(inner string) string {
	return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>RimixFac 授权</title>
<style>body{font-family:"Microsoft YaHei",sans-serif;background:#1a140c;color:#f3e6d0;display:flex;min-height:100vh;align-items:center;justify-content:center}
.card{max-width:520px;background:#241c14;padding:32px;border:1px solid #3d3224;line-height:1.6}</style></head>
<body><div class="card"><h2>RimixFac 永久授权</h2>` + inner + `</div></body></html>`
}
