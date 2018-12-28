package rtmp

const (
	rtmp_state_init         = 1
	rtmp_state_hand_fail    = 2
	rtmp_state_hand_success = 3

	rtmp_state_connect_send    = 10
	rtmp_state_connect_success = 11
	rtmp_state_connect_fail    = 1019

	rtmp_state_crtstm_send     = 20
	rtmp_state_crtstrm_success = 21
	rtmp_state_crtstrm_fail    = 1029

	rtmp_state_publish_send    = 30
	rtmp_state_publish_success = 31
	rtmp_state_publish_fail    = 1039

	rtmp_state_play_send           = 40
	rtmp_state_stream_is_record    = 41
	rtmp_state_stream_begin        = 42
	rtmp_state_play_reset          = 43
	rtmp_state_play_start          = 44
	rtmp_state_data_start          = 45
	rtmp_state_play_publish_notify = 46
	rtmp_state_play_fail           = 1049

	rtmp_state_stop
)
