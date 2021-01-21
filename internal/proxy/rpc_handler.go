package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/FZambia/viper-lite"
	"github.com/centrifugal/centrifuge"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
)

// RPCHandlerConfig ...
type RPCHandlerConfig struct {
	Proxy RPCProxy
}

// RPCHandler ...
type RPCHandler struct {
	config    RPCHandlerConfig
	summary   prometheus.Observer
	histogram prometheus.Observer
	errors    prometheus.Counter
	nc        *nats.Conn
}

// NewRPCHandler ...
func NewRPCHandler(c RPCHandlerConfig) *RPCHandler {
	v := viper.GetViper()

	url := v.GetString("pm_nats_url")
	var nc *nats.Conn
	if url != "" {
		var err error
		nc, err = nats.Connect(
			url,
			//nats.ReconnectBufSize(-1),
			//nats.MaxReconnects(-1),
			//nats.Timeout(time.Duration(viper.GetInt("pm_nats_dial_timeout")) * time.Second),
			//nats.FlusherTimeout(time.Duration(viper.GetInt("pm_nats_write_timeout")) * time.Second),
		)
		if err != nil {
			fmt.Println("nats error => ", err)
		}
	}
	return &RPCHandler{
		config:    c,
		summary:   proxyCallDurationSummary.WithLabelValues(c.Proxy.Protocol(), "rpc"),
		histogram: proxyCallDurationHistogram.WithLabelValues(c.Proxy.Protocol(), "rpc"),
		errors:    proxyCallErrorCount.WithLabelValues(c.Proxy.Protocol(), "rpc"),
		nc:        nc,
	}
}

// RPCHandlerFunc ...
type RPCHandlerFunc func(*centrifuge.Client, centrifuge.RPCEvent) (centrifuge.RPCReply, error)

// Handle RPC.
func (h *RPCHandler) Handle(node *centrifuge.Node) RPCHandlerFunc {
	return func(client *centrifuge.Client, e centrifuge.RPCEvent) (centrifuge.RPCReply, error) {
		//started := time.Now()
		reqData := RPCRtoNATSRequest{
			Method:    e.Method,
			Data:      string(e.Data),
			ClientID:  client.ID(),
			UserID:    client.UserID(),
			Transport: client.Transport(),
		}
		data, err := json.Marshal(reqData)
		if err != nil {
			fmt.Println("nats send err => ", err)
			return centrifuge.RPCReply{}, err
		}
		err = h.nc.Publish(e.Method, data)
		//duration := time.Since(started).Seconds()
		if err != nil {
			fmt.Println("nats send err 2 => ", err)
			return centrifuge.RPCReply{}, err
		}

		return centrifuge.RPCReply{}, nil

		//rpcRep, err := h.config.Proxy.ProxyRPC(client.Context(), RPCRequest{
		//	Method:    e.Method,
		//	Data:      e.Data,
		//	ClientID:  client.ID(),
		//	UserID:    client.UserID(),
		//	Transport: client.Transport(),
		//})
		//duration := time.Since(started).Seconds()
		//if err != nil {
		//	if errors.Is(err, context.Canceled) {
		//		return centrifuge.RPCReply{}, nil
		//	}
		//	h.summary.Observe(duration)
		//	h.histogram.Observe(duration)
		//	h.errors.Inc()
		//	node.Log(centrifuge.NewLogEntry(centrifuge.LogLevelError, "error proxying RPC", map[string]interface{}{"error": err.Error()}))
		//	return centrifuge.RPCReply{}, centrifuge.ErrorInternal
		//}
		//h.summary.Observe(duration)
		//h.histogram.Observe(duration)
		//if rpcRep.Disconnect != nil {
		//	return centrifuge.RPCReply{}, rpcRep.Disconnect
		//}
		//if rpcRep.Error != nil {
		//	return centrifuge.RPCReply{}, rpcRep.Error
		//}
		//
		//rpcData := rpcRep.Result
		//var data []byte
		//if rpcData != nil {
		//	if client.Transport().Encoding() == "json" {
		//		data = rpcData.Data
		//	} else {
		//		if rpcData.Base64Data != "" {
		//			decodedData, err := base64.StdEncoding.DecodeString(rpcData.Base64Data)
		//			if err != nil {
		//				node.Log(centrifuge.NewLogEntry(centrifuge.LogLevelError, "error decoding base64 data", map[string]interface{}{"client": client.ID(), "error": err.Error()}))
		//				return centrifuge.RPCReply{}, centrifuge.ErrorInternal
		//			}
		//			data = decodedData
		//		}
		//	}
		//}
		//
		//return centrifuge.RPCReply{
		//	Data: data,
		//}, nil
	}
}
