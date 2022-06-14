package demo

import (
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

type PostgresRequest struct {
	Username string `json:"username" desc:"Access User Name" binding:"required" faker:"username" `
}

type PostgresResponse struct {
	Code       int          `json:"code" desc:"0: success other: fail"`
	AccessLogs []AccessInfo `json:"accessLogs" desc:"Access record"`
}

type PostgresHandler struct {
	Request  PostgresRequest
	Response PostgresResponse
}

func (h PostgresHandler) Mock() interface{} {
	logs := make([]AccessInfo, 0, 10)
	for i := 0; i < 10; i++ {
		var info AccessInfo
		_ = faker.FakeData(&info)
		logs = append(logs, info)
	}

	return PostgresResponse{Code: 0, AccessLogs: logs}

}

func (h PostgresHandler) Handler(c *gateway.ApiContext) (interface{}, error) {

	log.Default().Info("Enter MysqlHandler")

	db := data.MustGetDB("hello")
	createSql := `
	CREATE TABLE if not exists access_log(
		id  SERIAL PRIMARY KEY,
		username           TEXT      NOT NULL,
		client_ip        TEXT,
		ts            INT       NOT NULL
	 )
	`

	err := db.Exec(createSql).Error
	util.CheckError(err)

	insertSql := "insert into access_log (username, client_ip, ts) values (?, ?, ?)"
	err = db.Exec(insertSql, h.Request.Username, c.ClientIP, time.Now().Unix()).Error
	util.CheckError(err)

	var logs []AccessInfo
	err = db.Table("access_log").Order("ts desc").Limit(10).Find(&logs).Error
	util.CheckError(err)

	return PostgresResponse{Code: 0, AccessLogs: logs}, nil
}
