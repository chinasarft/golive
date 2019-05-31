package signaling

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
	"github.com/satori/go.uuid"
)

type webrtcTransport struct {
	id       string
	reqParam *msCreateWebrtcTprReq
}

type mediasoupRouter struct {
	id         string
	room       *mediasoupRoom
	transports map[string]*webrtcTransport
	//producers map[string]string
}

type mediasoupRoom struct {
	roomId     string
	router     *mediasoupRouter
	audioObrId string
	forceH264  bool
	worker     *mediasoupWorker
	peers      sync.Map // map[peerid]*mediaPeer
}

type mediasoupPeer struct {
	peerId   string
	room     *mediasoupRoom
	joinRep  *msJoinReq
	isJoined bool
}

type mediasoupHouse struct { // house
	rooms    sync.Map // map[string]*mediasoupRoom
	director *mediasoupDirector
}

func newMansion(d *mediasoupDirector) *mediasoupHouse {
	return &mediasoupHouse{
		director: d,
	}
}

func (h *mediasoupHouse) getOrCreateRoom(roomId string, forceH264 bool) (*mediasoupRoom, error) {
	if v, ok := h.rooms.Load(roomId); ok {
		return v.(*mediasoupRoom), nil
	}

	w := h.director.getWorker()

	var err error
	var r *mediasoupRoom
	if r, err = createRoom(w, forceH264, roomId); err != nil {
		log.Error().Err(err).Msg("createRoom")
		return nil, err
	}
	h.rooms.Store(roomId, r)

	return r, nil
}

func createRoom(w *mediasoupWorker, forceH264 bool, roomId string) (*mediasoupRoom, error) {
	r := &mediasoupRoom{
		roomId:    roomId,
		forceH264: forceH264,
		worker:    w,
	}
	u, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	r.router = &mediasoupRouter{
		id:         u.String(),
		room:       r,
		transports: make(map[string]*webrtcTransport),
	}

	u, err = uuid.NewV4()
	if err != nil {
		return nil, err
	}
	r.audioObrId = u.String()

	reqId := getGlobalReqCounter()
	reqStr := fmt.Sprintf(`{"id":%d,"method":"worker.createRouter","internal":{"routerId":"%s"}}`,
		reqId, r.router.id)
	if _, err = w.ss.doRequest(w.ipc, reqId, reqStr); err != nil {
		return nil, err
	}

	reqId = getGlobalReqCounter()
	reqStr = fmt.Sprintf(`{"id":%d,"method":"router.createAudioLevelObserver",`+
		`"internal":{"routerId":"%s","rtpObserverId":"%s"},`+
		`"data":{"maxEntries":1,"threshold":-80,"interval":800}}`,
		reqId, r.router.id, r.audioObrId)
	if _, err = w.ss.doRequest(w.ipc, reqId, reqStr); err != nil {
		return nil, err
	}

	return r, nil
}

// mediasoup 的peer概念在protoo这一层，应该没有发送到worker的，暂时还不清楚怎么做
func (r *mediasoupRoom) createPeer() {

}

func (r *mediasoupRouter) createWebrtcTransport(conf *config.MediasoupConfig, cwtConf *msCreateWebrtcTprReq) (*msWebrtcTransportResp, error) {
	reqId := getGlobalReqCounter()
	u, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	tprId := u.String()

	listenIps, _ := json.Marshal(&conf.WebRtcTransport.ListenIps)
	reqStr := fmt.Sprintf(`{"id":%d,"method":"router.createWebRtcTransport",`+
		`"internal":{"routerId":"%s","transportId":"%s"},`+
		`"data":{"listenIps":`+string(listenIps)+
		`,"enableUdp":true,"enableTcp":true,"preferUdp":true,"preferTcp":false,`+
		`"initialAvailableOutgoingBitrate":%d,"minimumAvailableOutgoingBitrate":100000}}`,
		reqId, r.id, tprId, conf.WebRtcTransport.InitialAvailableOutgoingBitrate)

	var msg *msWorkerResp
	if msg, err = r.room.worker.ss.doRequest(r.room.worker.ipc, reqId, reqStr); err != nil {
		return nil, err
	}

	resp := &msWebrtcTransportResp{}
	if err = json.Unmarshal(msg.Data, resp); err != nil {
		log.Error().Err(err).Msg("unmarshal")
		return nil, err
	}
	r.transports[tprId] = &webrtcTransport{
		id:       tprId,
		reqParam: cwtConf,
	}

	return resp, nil
}

func (r *mediasoupRouter) join(joinReq *msJoinReq) (*msJoinResp, error) {
	// 返回除自己之外的其它peer
	return &msJoinResp{Peers: make([]string, 0)}, nil
}

func (r *mediasoupRouter) connectWebRtcTransport(connReq *msConnectRtcTprReq) (*msConnectRtcTprResp, error) {

	reqId := getGlobalReqCounter()

	// {"id":7,"method":"transport.connect","internal":{"routerId":"9f12f40d-f4f5-402c-b33f-c08aeaadffac",
	//"transportId":"a14e6985-f757-47fa-b952-f181d0b64286"},"data":{"dtlsParameters":{"role":"server",
	//"fingerprints":[{"algorithm":"sha-256","value":"6B:6B:02:23:41:1E:3C:13:01:BA:3B:A2:EF:43:39:68:AA:C4:E6:02:93:EE:9D:10:7A:79:74:A0:F2:15:72:E6"}]}}}

	reqStr := fmt.Sprintf(`{"id":%d,"method":"transport.connect","internal":{"routerId":"%s`+
		`","transportId":"%s"},"data":{"dtlsParameters":{"role":"%s","fingerprints":`+
		`[{"algorithm":"%s","value":"%s"}]}}}`, reqId, r.id, connReq.TransportId, connReq.DtlsParameters.Role,
		connReq.DtlsParameters.Fingerprints[0].Algorithm, connReq.DtlsParameters.Fingerprints[0].Value)
	var msg *msWorkerResp
	var err error
	if msg, err = r.room.worker.ss.doRequest(r.room.worker.ipc, reqId, reqStr); err != nil {
		return nil, err
	}
	resp := &msConnectRtcTprResp{}
	if err = json.Unmarshal(msg.Data, resp); err != nil {
		log.Error().Str("str", string(msg.Data)).Msg("")
		log.Error().Err(err).Msg("unmarshal")
		return nil, err
	}
	log.Debug().Str("connect", "role").Msg(resp.DtlsLocalRole)

	return &msConnectRtcTprResp{}, nil
}

func (r *mediasoupRouter) produce(produceReq *msProduceReq) (*msProduceResp, error) {
	return nil, nil
}
