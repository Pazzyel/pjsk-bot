package main

import (
	"log"
	"pjsk-bot/bot/config"
	"pjsk-bot/bot/event"
	"pjsk-bot/bot/qqclient"
	"strconv"
)

func main() {
	qqConfig, err := config.LoadConfig("bot/resources")
	if err != nil {
		log.Fatal("加载配置文件失败\n", err)
	}
	qqNumber, err := strconv.Atoi(qqConfig.QQ.Number)
	if err != nil {
		log.Fatal("无法将QQ号转换为uint32")
	}
	myClient := qqclient.NewMyClient(uint32(qqNumber), qqConfig.QQ.Password)
	myEvent := event.NewEvent(myClient)
	myEvent.Start()
	for true {

	}
}
