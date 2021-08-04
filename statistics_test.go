package workerman_statistics_go

import (
	"testing"
	"time"
)

func TestDEncode(t *testing.T) {
	client := NewWorkerManClient("127.0.0.1", 0, "127.0.0.1", 0)
	infoX := WorkerManMsgInfo{Module: "test", InterFace: "api", CostTime: 10.01, Code: 200, Status: 1, TimeStamp: uint32(time.Now().Unix()), Msg: "err"}
	bytes, err := client.Encode(infoX)
	if err != nil {
		t.Error(err)
		return
	}
	info2 := client.Decode(bytes)
	t.Logf("%+v", info2)
}
