// Package mockpsp2 提供第二条内存 mock 上游通道 mock_psp_alt，与 mock_psp 并行用于联调。
//
// 与 mock_psp 的差异（约定）：
//   - JSON 字段为 snake_case：merchant_ref、txn_id、state、amount、event_time、signature（代收）；
//     代付另含 payout_state、bank_reference。
//   - state / payout_state 使用英文枚举：PENDING、SUCCESS、FAIL（代收）；PROCESSING、SUCCESS、FAIL（代付）。
//   - 签名为 MD5（键名小写、排序、非空值、&key=secret），见 SignMd5SortedKV；非 mock_psp 的 HMAC-SHA256。
//   - 应答上游仍为纯文本 SUCCESS/FAIL（与 gateway checkout 回调一致）。
//
// 注册：channels.payin_type = mock_psp_alt；sign_secret 与库表一致。
package mockpsp2
