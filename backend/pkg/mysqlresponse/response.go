package mysqlresponse

import "github.com/gin-gonic/gin"

type JSONResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(200, JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func Error(ctx *gin.Context, code int, msg string) {
	ctx.JSON(code, JSONResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
