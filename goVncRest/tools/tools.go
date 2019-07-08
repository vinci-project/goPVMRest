package tools

import (
	"github.com/valyala/fasthttp"
)

func MakeResponse(statusCode int,
	ctx *fasthttp.RequestCtx) {
	//

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(statusCode)
}
