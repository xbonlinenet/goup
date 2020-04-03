package demo

import (
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

type MysqlRequest struct {
	Username string `json:"username" desc:"Access User Name" binding:"required" faker:"username" `
}

type AccessInfo struct {
	Id         int    `json:"id" gorm:"column:id" desc:"Identify"`
	Username   string `json:"username" gorm:"column:username" desc:"Access Username"`
	ClientIP   string `json:"clientIP" gorm:"column:client_ip" desc:"Where user access"`
	AccessUnix int64  `json:"accessUnix" gorm:"column:ts" desc:"When user access"`
}

type MysqlResponse struct {
	Code       int          `json:"code" desc:"0: success other: fail"`
	AccessLogs []AccessInfo `json:"accessLogs" desc:"Access record"`
}

type MysqlHandler struct {
	Request  MysqlRequest
	Response MysqlResponse
}

func (h MysqlHandler) Mock() interface{} {

	logs := make([]AccessInfo, 0, 10)
	for i := 0; i < 10; i++ {
		var info AccessInfo
		_ = faker.FakeData(&info)
		logs = append(logs, info)
	}

	return MysqlResponse{Code: 0, AccessLogs: logs}

}

func (h MysqlHandler) Handler(c *gateway.ApiContext) (interface{}, error) {

	log.Default().Info("Enter MysqlHandler")

	db := data.MustGetDB("demo")
	createSql := `
	create table if not Exists goup.access_log  (
		id int(11) NOT NULL AUTO_INCREMENT,
		username varchar(128) NOT NULL COMMENT '用户名',
		client_ip varchar(128) NOT NULL COMMENT 'IP 地址',
		ts varchar(128) NOT NULL COMMENT '访问时间',
		PRIMARY KEY (id),
		KEY idx_ts (ts)
	)
	`

	err := db.Exec(createSql).Error
	util.CheckError(err)

	insertSql := "insert into goup.access_log (username, client_ip, ts) values (?, ?, ?)"
	err = db.Exec(insertSql, h.Request.Username, c.ClientIP, time.Now().Unix()).Error
	util.CheckError(err)

	var logs []AccessInfo
	err = db.Table("goup.access_log").Order("ts desc").Limit(10).Find(&logs).Error
	util.CheckError(err)

	return MysqlResponse{Code: 0, AccessLogs: logs}, nil
}
