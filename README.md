# RimixFac 授权支付后端

Vercel Serverless（Go）+ 易支付（Go 码）+ Vercel KV。

密钥只放 Vercel 环境变量，不要提交到 Git。

## 接口

- `POST /api/getPayUrl` 下单，返回微信付款链接
- `POST /api/notify` Go 码异步回调（必须回复 `success`）
- `GET /api/paid` 支付完成展示卡密
- `POST /api/activate` 卡密 + 设备指纹激活（首次绑定，允许换机 1 次）

套餐：`rimix-perm-399` 或 `ace-perm-399`，默认 399 元永久。

## 环境变量

| 变量 | 说明 |
| --- | --- |
| GOPAY_PID | 商户 ID |
| GOPAY_SECRET | 商户密钥（在 Go 码后台，勿发到聊天） |
| GOPAY_SUBMIT_URL | 可选，默认 `https://pay.hunyuantaiji.shop/xpay/epay/submit.php` |
| API_TOKEN | Rimix 调用接口的 Bearer Token |
| KV_REST_API_URL | Vercel KV / Upstash REST URL |
| KV_REST_API_TOKEN | Vercel KV REST Token |
| PACKAGE_PRICE | 可选，默认 `399.00` |
| PACKAGE_NAME | 可选，默认 `RimixFac永久授权` |

`notify_url` / `return_url` 由代码按当前域名自动生成，不必先填占位再改。

## 鉴权

`getPayUrl`、`activate` 请求头：

`Authorization: Bearer <API_TOKEN>`
