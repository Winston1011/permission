// Package nmq 提供通过边车访问 NMQ 的能力
package nmq

type NmqResponse struct {
	TransID uint64 `mcpack:"_transid"`
	ErrNo   int    `mcpack:"_error_no" binding:"required"`
	ErrStr  string `mcpack:"_error_msg" binding:"required"`
}
