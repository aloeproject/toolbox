package response

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
)

type Response struct {
	Code      int32       `json:"code"`
	Msg       string      `json:"msg"`
	RequestId string      `json:"request_id"`
	Time      time.Time   `json:"time"`
	Data      interface{} `json:"data"`
}

type IResponse interface {
	JSON(code int, obj interface{})
}

func NewHttpResponse(requestId string) *HttpResponse {
	return &HttpResponse{
		requestId: requestId,
	}
}

type HttpResponse struct {
	iRes      IResponse
	ctx       context.Context
	requestId string
}

func (h *HttpResponse) ResponseWithMessage(httpCode, errCode int32, data interface{}, msg string) {
	var requestId string
	if h.requestId == "" {
		if span := trace.SpanContextFromContext(h.ctx); span.HasTraceID() {
			requestId = span.TraceID().String()
		}
	}

	response := Response{
		Code:      errCode,
		Msg:       msg,
		Data:      data,
		Time:      time.Now(),
		RequestId: requestId,
	}

	h.iRes.JSON(int(httpCode), response)
	return
}

func (h *HttpResponse) InitResponse(ctx context.Context, iRes IResponse) {
	h.ctx = ctx
	h.iRes = iRes
}

func (h *HttpResponse) InitByGin(ctx *gin.Context) {
	h.ctx = ctx.Request.Context()
	h.iRes = ctx
}

func (h *HttpResponse) ResponseSuccess(data interface{}) {
	h.ResponseWithMessage(http.StatusOK, 200, data, "")
}

func (h *HttpResponse) ResponseError(errCode int32, data interface{}, msg string) {
	h.ResponseWithMessage(http.StatusOK, errCode, data, msg)
}
