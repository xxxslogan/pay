package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xxxslogan/pay/internal"
)

type activateReq struct {
	CardKey  string `json:"cardKey"`
	DeviceID string `json:"deviceId"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if internal.CORS(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		internal.JSON(w, 200, map[string]any{"code": 405, "msg": "method not allowed"})
		return
	}
	if !internal.CheckAPIToken(r) {
		internal.JSON(w, 200, map[string]any{"code": 401, "msg": "unauthorized"})
		return
	}
	var req activateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.JSON(w, 200, map[string]any{"code": 400, "msg": "invalid json"})
		return
	}
	req.CardKey = strings.TrimSpace(req.CardKey)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.CardKey == "" || req.DeviceID == "" {
		internal.JSON(w, 200, map[string]any{"code": 400, "msg": "missing cardKey or deviceId"})
		return
	}
	var card internal.Card
	if err := internal.GetJSON("card:"+req.CardKey, &card); err != nil {
		internal.JSON(w, 200, map[string]any{"code": 1001, "msg": "卡密不存在"})
		return
	}
	if card.BindDevice == "" {
		card.BindDevice = req.DeviceID
		if err := internal.SetJSON("card:"+req.CardKey, card); err != nil {
			internal.JSON(w, 200, map[string]any{"code": 500, "msg": "bind fail"})
			return
		}
		internal.JSON(w, 200, map[string]any{
			"code": 0, "msg": "激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
		})
		return
	}
	if card.BindDevice == req.DeviceID {
		internal.JSON(w, 200, map[string]any{
			"code": 0, "msg": "激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
		})
		return
	}
	if card.ChangeLeft <= 0 {
		internal.JSON(w, 200, map[string]any{"code": 1002, "msg": "换机次数耗尽"})
		return
	}
	card.BindDevice = req.DeviceID
	card.ChangeLeft--
	if err := internal.SetJSON("card:"+req.CardKey, card); err != nil {
		internal.JSON(w, 200, map[string]any{"code": 500, "msg": "rebind fail"})
		return
	}
	internal.JSON(w, 200, map[string]any{
		"code": 0, "msg": "换机激活成功", "expireTime": 0, "licenseType": "永久", "changeLeft": card.ChangeLeft,
	})
}
