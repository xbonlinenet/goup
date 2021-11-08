package gateway

import (
	"github.com/gin-gonic/gin"
)

//解密处理器,用于解析request前解密request数据
type CryptoHandler struct {
	decryptImpl func(c *gin.Context) bool  //数据解密
	encryptImpl func(d interface{}) string //返回结果加密
}

func NewCryptoHandler(encryptImpl func(d interface{}) string, decryptImpl func(c *gin.Context) bool) *CryptoHandler {
	return &CryptoHandler{
		decryptImpl: decryptImpl,
		encryptImpl: encryptImpl,
	}
}
func (crypto *CryptoHandler) Decrypt(c *gin.Context) bool {
	return crypto.Decrypt(c)
}
func (crypto *CryptoHandler) Encrypt(d interface{}) string {
	return crypto.encryptImpl(d)
}
func WithCryptoHandler(h *CryptoHandler) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.cryptoHandler = h
	})
}
