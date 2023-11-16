package helpers

import (
	"permission/pkg/golib/v2/kms"
)

var Kms kms.Kms

func InitKms() {
	Kms = kms.Init()
}
