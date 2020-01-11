package pwd

import (
	"github.com/common/logger"
)

var Encrypter PwdEncrypter

type PwdEncrypter interface {
	Encrypter(password string) (string, error)
	Decrypter(code string) (string, error)
}

func NewPwdEncrypter(encrypter string) {
	switch encrypter {
	case "aes":
		Encrypter = NewAES()
	default:
		logger.Info("Use default encryption tool: aes.")
		Encrypter = NewAES()
	}
}
