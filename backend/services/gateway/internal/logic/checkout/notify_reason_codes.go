package logic

// Upstream callback reason_code (stable machine-readable values).
const (
	NotifyCodeInvalidNotifyParams   = "INVALID_NOTIFY_PARAMS"
	NotifyCodeChannelNotFound       = "CHANNEL_NOT_FOUND"
	NotifyCodeInvalidSign           = "INVALID_SIGN"
	NotifyCodeOrderNotFound         = "ORDER_NOT_FOUND"
	NotifyCodeOrderNotPending       = "ORDER_NOT_PENDING"
	NotifyCodeReplayPayloadMismatch = "REPLAY_PAYLOAD_MISMATCH"
	NotifyCodeMarkPaidFailed        = "MARK_PAID_FAILED"
	NotifyCodeMarkPaidRace          = "MARK_PAID_RACE"
	NotifyCodeMarkPaidRaceMismatch  = "MARK_PAID_RACE_MISMATCH"

	NotifyCodeIdempotentReplayAccepted = "IDEMPOTENT_REPLAY_ACCEPTED"
	NotifyCodeIdempotentRaceAccepted   = "IDEMPOTENT_RACE_ACCEPTED"
)
