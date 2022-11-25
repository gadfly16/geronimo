package main

type broker struct {
	id        int64
	accountId int64
	name      string
	status    string
	base      float64
	quote     float64
	minWait   float64
	maxWait   float64
	highLimit float64
	lowLimit  float64
	delta     float64
	offset    float64
}

// func getBroker(name string) *broker {
// 	db := openDB()
// 	defer db.Close()

// 	sqlStmt = `
// 		SELECT * FROM broker b
// 		JOIN brokerSetting bs ON b.id=bs.broker
// 		JOIN brokerBalance bb ON b.id=bb.broker ;
// 	`
// }
