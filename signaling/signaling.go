package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
	"github.com/chinasarft/golive/signaling/protoo"
	"github.com/gorilla/websocket"
)

type signalingServer struct {
	server    http.Server
	mux       *http.ServeMux
	isStarted bool
	conf      config.SigalingConfig
	director  *mediasoupDirector
	houses    *mediasoupHouse
}

var server signalingServer

// websocket: request origin not allowed by Upgrader.CheckOrigin
// TODO 这个错误需要解决， 暂时先这样解决
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getSingleQuery(name string, q url.Values) string {
	if _, ok := q[name]; ok {
		return q[name][0]
	}
	return ""
}

func protooHandler(w http.ResponseWriter, r *http.Request) {

	swp := r.Header.Get("Sec-WebSocket-Protocol")
	if swp != "" {
		w.Header().Set("Sec-WebSocket-Protocol", swp)
	}
	c, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Error().Err(err).Msg("upgrade websocket fail")
		return
	}

	defer c.Close()

	qs := r.URL.Query()
	roomId := getSingleQuery("roomId", qs)
	peerId := getSingleQuery("peerId", qs)
	forceH264 := getSingleQuery("forceH264", qs) == "true"

	if roomId == "" || peerId == "" {
		log.Info().Msg("no roomId or peerId")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Connection request without roomId and/or peerId"))
		return
	}
	log.Info().Str("roomId", roomId).Str("peerId", peerId).Bool("forceH264", forceH264).Msg("connected")
	var R *mediasoupRoom
	defer func() {
		if R != nil {
			if _, ok := R.peers.Load(peerId); ok {
				R.peers.Delete(peerId)
			}
		}
	}()
	for {
		req := protoo.Request{}

		msgType, message, err := c.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("read websocket")
			break
		}
		log.Debug().Msg(string(message))
		if msgType != websocket.TextMessage {
			log.Error().Int("type", msgType).Msg("not text message")
			if err = req.ErrorResponse(c, "not text message"); err != nil {
				log.Error().Err(err).Msg("response fail")
			}
			break
		}

		if err = json.Unmarshal(message, &req); err != nil {
			log.Error().Err(err).Str("body", string(message)).Msg("wrong request")
			if err = req.ErrorResponse(c, "wrong request"); err != nil {
				log.Error().Err(err).Msg("response fail")
			}
			break
		}

		R, err = server.houses.getOrCreateRoom(roomId, false)
		if err != nil {
			req.ErrorResponse(c, "createroom")
			continue
		}

		// node里面用这个peer去拿websocket的链接，但是go里面不用这样去拿
		R.peers.Store(peerId, R)

		var respData interface{}
		switch req.Method {
		case "getRouterRtpCapabilities":
			respData = &expectCaps
		case "createWebRtcTransport":
			cwtReq := &msCreateWebrtcTprReq{}
			if err = json.Unmarshal(req.Data, cwtReq); err == nil {
				respData, err = R.router.createWebrtcTransport(&server.director.conf, cwtReq)
			}
		case "join":
			joinReq := &msJoinReq{}
			if err = json.Unmarshal(req.Data, joinReq); err == nil {
				respData, err = R.router.join(joinReq)
			}
		case "connectWebRtcTransport":
			connReq := &msConnectRtcTprReq{}
			if err = json.Unmarshal(req.Data, connReq); err == nil {
				respData, err = R.router.connectWebRtcTransport(connReq)
			}

		case "produce":
			produceReq := &msProduceReq{}
			if err = json.Unmarshal(req.Data, produceReq); err == nil {
				respData, err = R.router.produce(produceReq)
			}
		}
		if err == nil {
			d := req.GetResponseData(respData)
			if err = c.WriteMessage(msgType, d); err != nil {
				log.Error().Err(err).Msg("response fail")
			}
		} else {
			if err = req.ErrorResponse(c, err.Error()); err != nil {
				log.Error().Err(err).Msg("response fail")
			}
		}
	}
}

func StartProtooSignaling(conf *config.SigalingConfig) (err error) {
	if server.isStarted {
		return fmt.Errorf("already started")
	}

	if err = setExpectCapsFromConfigAndSupportedCaps(conf.Mediasoup.Router.MediaCodecs); err != nil {
		return err
	}

	if server.director, err = newMediasoupDirector(&conf.Mediasoup); err != nil {
		return err
	}
	server.houses = newMansion(server.director)

	server.isStarted = true
	server.server.Addr = conf.Protoo.Addr
	server.installHander()
	server.conf = *conf

	go server.start()
	return nil
}

func (s *signalingServer) start() {
	if s.conf.Protoo.Cert != "" && s.conf.Protoo.Key != "" {
		if err := s.server.ListenAndServeTLS(s.conf.Protoo.Cert, s.conf.Protoo.Key); err != nil {
			log.Error().Err(err).Msg("admin finished")
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("admin finished")
		}
	}
}

func (s *signalingServer) installHander() {

	mux := http.NewServeMux()
	server.mux = mux
	s.server.Handler = mux

	mux.Handle("/", http.HandlerFunc(protooHandler))
}

func StopProtooSignaling() error {

	if server.isStarted {
		ctx, _ := context.WithCancel(context.Background())
		if err := server.server.Shutdown(ctx); err != nil {
			return err
		}
		server.isStarted = false
	}
	return nil
}
