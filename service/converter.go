package service

import (
	"errors"
	"fmt"
	"github.com/channingduan/rpc/config"
	"github.com/smallnest/rpcx/codec"
	"github.com/smallnest/rpcx/protocol"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	XVersion           = "X-RPCX-Version"
	XMessageType       = "X-RPCX-MesssageType"
	XHeartbeat         = "X-RPCX-Heartbeat"
	XOneway            = "X-RPCX-Oneway"
	XMessageStatusType = "X-RPCX-MessageStatusType"
	XSerializeType     = "X-RPCX-SerializeType"
	XMessageID         = "X-RPCX-MessageID"
	XServicePath       = "X-RPCX-ServicePath"
	XServiceMethod     = "X-RPCX-ServiceMethod"
	XMeta              = "X-RPCX-Meta"
	XErrorMessage      = "X-RPCX-ErrorMessage"
)

// HttpRequest2RPCRequest 协议转换
func HttpRequest2RPCRequest(r *http.Request) (*protocol.Message, error) {

	req := protocol.NewMessage()
	req.SetMessageType(protocol.Request)
	h := r.Header
	seq := h.Get(XMessageID)
	if seq != "" {
		id, err := strconv.ParseUint(seq, 10, 64)
		if err != nil {
			return nil, err
		}
		req.SetSeq(id)
	}

	fmt.Println("r.Header.Get(XServicePath)1", r.Header.Get(XServicePath))
	fmt.Println("r.Header.Get(XServiceMethod)1", r.Header.Get(XServiceMethod))

	heartbeat := h.Get(XHeartbeat)
	if heartbeat != "" {
		req.SetHeartbeat(true)
	}
	oneway := h.Get(XOneway)
	if oneway != "" {
		req.SetOneway(true)
	}
	if h.Get("Content-Encoding") == "gzip" {
		req.SetCompressType(protocol.Gzip)
	}
	serializeType := h.Get(XSerializeType)
	if serializeType != "" {
		rst, err := strconv.Atoi(serializeType)
		if err != nil {
			return nil, err
		}
		req.SetSerializeType(protocol.SerializeType(rst))
	} else {
		return nil, errors.New("empty serialized type")
	}

	meta := h.Get(XMeta)
	if meta != "" {
		metadata, err := url.ParseQuery(meta)
		if err != nil {
			return nil, err
		}
		mm := make(map[string]string)
		for k, v := range metadata {
			if len(v) > 0 {
				mm[k] = v[0]
			}
		}
		req.Metadata = mm
	}

	servicePath := h.Get(XServicePath)
	if servicePath != "" {
		req.ServicePath = servicePath
	} else {
		return nil, errors.New("empty service path")
	}

	serviceMethod := h.Get(XServiceMethod)
	if serviceMethod != "" {
		req.ServiceMethod = serviceMethod
	} else {
		return nil, errors.New("empty service method")
	}
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// 编码
	cc := &codec.MsgpackCodec{}
	args := config.Request{
		Message: string(payload),
	}
	req.Payload, _ = cc.Encode(args)

	return req, nil
}
