package segment

type GameSegmentReq struct {
	TableId       int           `json:"table_id"`
	LotteryRecord []interface{} `json:"lottery_record"`
	UserCount     int           `json:"user_count"`
	JettonTime    int           `json:"jetton_time"`
	LotteryTime   int           `json:"lottery_time"`
	ResidueTime   int           `json:"residue_time"`
	RoomStatus    int           `json:"room_status"`
	JettonList    []float32     `json:"jetton_list"`
	Astrict       float32       `json:"astrict"`
}
