package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xxxslogan/pay/pkg"
)

type activateReq struct {
	CardKey  string `json:"cardKey"`
	DeviceID string `json:"deviceId"`
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
