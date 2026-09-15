package handler

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xxxslogan/pay/pkg"
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
