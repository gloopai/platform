# 印度上游 PSP 对接说明

本目录描述**一家**印度四方/聚合支付上游的 HTTP API（商户侧调上游）。平台内部实现上游通道时，应以本文为契约；**不要求**与 pay-platform 对商户开放的 OpenAPI 路径或字段名一致。
接入须知
约 7 字
小于 1 分钟

收付款人注意事项

创建代收/代付中的姓名，手机，邮箱没有真实的话随机，不要固定写死,随机邮箱用gmail(这个很重要,写死的极易触发风控)

回调测试
回调

以下规则只在测试状态下生效

接口请求成功后将会在60秒内自动回调成功
若需要回调失败，请将name字段的值传为FAILURE。
代收需打开收银台才会回调
金额单位
金额

接口中所有金额单位都为分
网关地址
国家	协议	域名	回调IP	描述
印度	HTTPS	api.hexmeta.xyz	52.66.62.39	支付网关地址，请根据需要添加回调IP白名单
构造请求
请求URI

<protcol>://<domain>/<requestPath>
参数说明:

参数	必选	描述
protocol	是	请求使用的协议，如HTTP, HTTPS。HTTPS表示通过的安全的HTTPS访问该资源,请参考网关地址中的协议字段
domain	是	网关域名,请参考网关地址中的域名字段
requestPath	是	请求路径，具体可参考对应接口文档中的请求路径字段
请求方法
方法	说明
POST	本文档中所有请求都是以POST请求方式发出
请求头
提示

这两个请求头非常重要，必须要携带

请求头	必填	值	描述
Content-Type	是	application/json	必须要携带此参数，不然会出现appId not set的错误
Content-Length	是	body长度	请检查框架是否自动携带了此参数，有些语言的框架不会自动携带此参数
请求消息体
具体请求参数请参考对应接口的请求参数

公共请求参数示例
参数名	必填	类型	示例	描述
appId	是	string	023213567912	应用ID,创建商户后会提供
timestamp	是	string	1761118332428	unix(毫秒13位)时间戳, 和我方服务器时间误差不可超过5分钟
sign	是	string	022af4023d2ba3e197c5e2162e87b9fc	请求签名, 请参考签名规则
请求签名
注意

参数大小写敏感
值为空的参数不参与签名
ASCII码从小到大排序
签名结果为MD5签名后的16进制小写字符串
获取请求参数集合requestMap


{
  "orderNo": "9c576c2e-26f2-4bde-96a7-cf14c264e15b",
  "amount": "10000",
  "name": "name",
  "phone": "7277528013",
  "email": "Djhhkevi@example.xyz",
  "notifyUrl": "",
  "timestamp": "1761120765563",
  "appId": "023213567912"
}
拼接待签名字符串

对集合requestMap中非空参数值按照参数名ASCII码从小到大排序(字典序)
使用URL键值对格式key1=value1&key2=value2&key3=value3拼接成字符串requestParamsStr,待签名串应遵循如下规范:
参数名ASCII码从小到大排序
如果参数的值为空不参与签名
参数名区分大小写
接口可能增加或减少字段，签名时须支持增加或减少的字段
请求参数中必须包含参数timestamp且不为空，值为当前unix毫秒(13位)时间戳，该值时间不能和服务器时间相差大于5分钟，该参数必须参与签名
拼接密钥

将得到的待签名的字符串requestParamsStr末尾加&key={secret}(secret可在商户后台->基本信息->开发信息->更新密钥中获取), 最终得到待签名字符串signStr

amount=10000&appId=023213567912&email=Djhhkevi@example.xyz&name=name&orderNo=9c576c2e-26f2-4bde-96a7-cf14c264e15b&phone=7277528013&timestamp=1761120765563&key=ycqXEhpZIuZx1JV8yZem9V2I0NA2is0u
签名

对signStr进行MD5签名得到十六进制签名值sign(小写字符串),如:2ef0681c8984dc524d4ef203dacbc31a

将签名sign放入请求参数集合中，并发往服务器


{
  "orderNo": "9c576c2e-26f2-4bde-96a7-cf14c264e15b",
  "amount": "10000",
  "name": "name",
  "phone": "7277528013",
  "email": "Djhhkevi@example.xyz",
  "notifyUrl": "",
  "timestamp": "1761120765563",
  "appId": "023213567912",
  "sign": "2ef0681c8984dc524d4ef203dacbc31a"
}
请求返回
公共返回参数示例
参数	必带	类型	描述
code	是	int	请求错误码， 1-成功, 其他-失败
msg	是	string	错误原因, OK-成功, 其他-失败原因
data	否	object	响应数据,code为非1时不返回此字段,具体响应数据可参考具体接口的返回参数


---

## 1. 通用约定

- **Base path**：`/exposed/v1`（具体域名由上游提供）。
- **请求头**：`Content-Type: application/json`。
- **公共参数**（除回调外多数接口需要）：


| 参数          | 必填  | 说明                                  |
| ----------- | --- | ----------------------------------- |
| `appId`     | 是   | 应用 ID，开户后提供                         |
| `timestamp` | 是   | Unix 毫秒时间戳（13 位），与上游服务器时间误差不超过 5 分钟 |
| `sign`      | 是   | 请求签名，规则见上游《签名规则》文档（原文未附于本仓库）        |


- **响应信封**：


| 字段     | 类型     | 说明                |
| ------ | ------ | ----------------- |
| `code` | int    | `1` 表示成功，其他为失败    |
| `msg`  | string | `OK` 表示成功，其他为错误说明 |
| `data` | object | 业务数据              |


---

## 2. 代收（Payin）

### 2.1 创建代收订单

- **路径**：`POST /exposed/v1/order/payment`


| 参数          | 必填  | 类型     | 说明             |
| ----------- | --- | ------ | -------------- |
| `orderNo`   | 是   | string | 商户订单号，商户内唯一    |
| `amount`    | 是   | string | 金额，单位分，无小数     |
| `name`      | 是   | string | 付款人姓名，建议真实     |
| `phone`     | 是   | string | 付款人手机          |
| `email`     | 是   | string | 付款人邮箱          |
| `userIP`    | 否   | string | 用户 IP          |
| `notifyUrl` | 否   | string | 订单状态变更时的异步通知地址 |


**成功时 `data`：**


| 字段           | 说明    |
| ------------ | ----- |
| `sysOrderNo` | 平台订单号 |
| `payUrl`     | 收银台地址 |


### 2.2 查询代收订单

- **路径**：`POST /exposed/v1/query/payment`


| 参数        | 必填  | 说明    |
| --------- | --- | ----- |
| `orderNo` | 是   | 商户订单号 |


**成功时 `data` 含**：`appId`、`orderNo`、`sysOrderNo`、`amount`（分，string）、`status`、`referenceNo`、`failReason` 等。

`**status`**：`1` 处理中，`2` 成功，`3` 失败。

### 2.3 代收补单

- **路径**：`POST /exposed/v1/makeup`


| 参数            | 必填  | 说明    |
| ------------- | --- | ----- |
| `orderNo`     | 是   | 商户订单号 |
| `referenceNo` | 是   | UTR   |


若调用失败，可将补单截图与订单号发到 TG 商户群，由机器人协助补单追踪。

### 2.4 代收异步回调（上游 → 商户）

- **URL**：创建代收订单时传入的 `notifyUrl`。
- **方法**：`POST`，`Content-Type: application/json`。


| 参数           | 必填  | 说明                    |
| ------------ | --- | --------------------- |
| `timestamp`  | 是   | Unix 毫秒时间戳            |
| `sign`       | 是   | 签名                    |
| `orderNo`    | 是   | 商户订单号                 |
| `sysOrderNo` | 是   | 平台订单号                 |
| `status`     | 是   | `1` 处理中，`2` 成功，`3` 失败 |
| `amount`     | 是   | 实收金额，单位分（string）      |


**响应**：HTTP 200，body 为纯文本，**仅** `SUCCESS` 或 `FAIL`（忽略大小写）。非 200 或非上述字符串会触发重试（共 6 次，间隔递增）。可能重复回调，须幂等；重复处理成功时仍应返回 `SUCCESS`。

---

## 3. 代付（Payout）

### 3.1 创建代付订单

- **路径**：`POST /exposed/v1/order/payout`

**资损防范**：代付下单若失败，须先调**查询代付订单**确认实际状态；若查询不到（须处理超时、反序列化失败）或状态为失败，再换通道重试。


| 参数                         | 必填  | 说明                                     |
| -------------------------- | --- | -------------------------------------- |
| `orderNo`                  | 是   | 商户订单号，商户内唯一                            |
| `wayCode`                  | 是   | `1` 银行卡，`2` UPI；文档写明当前仅支持 `1`          |
| `amount`                   | 是   | 金额，单位分（string）                         |
| `bankName`                 | 是   | 开户银行名称，无则固定 `IndiaBank`                |
| `bankCode`                 | 是   | IFSC（四位大写 + `0` + 六位数字）                |
| `accountNo`                | 是   | `wayCode=1` 为银行卡号，`wayCode=2` 为 UPI 账号 |
| `name` / `phone` / `email` | 是   | 收款人信息，建议真实                             |
| `notifyUrl`                | 否   | 状态变更异步通知                               |


**成功时 `data.sysOrderNo`**：平台订单号。

### 3.2 查询代付订单

原文档未提供，**待上游补充**（路径、参数、状态枚举应对齐代收查询风格）。

### 3.3 代付异步回调（上游 → 商户）

- **URL**：创建代付订单时的 `notifyUrl`。
- **方法**：`POST`，`Content-Type: application/json`。


| 参数            | 必填  | 说明                    |
| ------------- | --- | --------------------- |
| `orderNo`     | 是   | 商户订单号                 |
| `sysOrderNo`  | 是   | 平台订单号                 |
| `status`      | 是   | `1` 处理中，`2` 成功，`3` 失败 |
| `amount`      | 是   | 交易金额，单位分（string）      |
| `referenceNo` | 是   | UTR                   |
| `timestamp`   | 是   | Unix 毫秒时间戳            |
| `sign`        | 是   | 签名                    |


**响应与重试、幂等**：同「代收异步回调」（返回 `SUCCESS` / `FAIL` 纯文本，6 次重试等）。

---

## 4. 查询商户余额

- **路径**：`POST /exposed/v1/query/balance`

仅需公共参数 `appId`、`timestamp`、`sign`。

**成功时 `data`：**


| 字段                 | 说明                 |
| ------------------ | ------------------ |
| `availableBalance` | 可用于代付的金额（分，string） |
| `unsettledAmount`  | 待结算（代收成功但尚不可代付）（分） |
| `frozenAmount`     | 冻结（代付处理中）（分）       |


---

