package antispam

import "errors"

var (
	ErrorParseParams        = errors.New("parse params error")
	ErrorAntiSpamSignErr    = errors.New("anti spam sign error")
	ErrorAntiSpamLackSign   = errors.New("user params lack sign key words")
	ErrorEmptyToken         = errors.New("empty token")
	ErrorToken              = errors.New("get broken token in redis")
	ErrorGetToken           = errors.New("get token in antispam-server failed")
	ErrorIgnore             = errors.New("ignore error")
	ErrorTokenNearlyExpired = errors.New("token nearly expired")
	ErrorTokenRc4           = errors.New("token rc4 error")
	ErrorCuIDEmpty          = errors.New("cuid is empty")
	ErrorLackTimeParam      = errors.New("lack time param: _t_")
	ErrorClientTimeFormat   = errors.New("client time format error")
	ErrorClientTimeTooOld   = errors.New("client time too old")
)
