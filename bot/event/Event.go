package event

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"pjsk-bot/bot/qqclient"
	"strconv"
	"strings"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/message"
)

type Event struct {
	myClient *qqclient.MyClient
}

func NewEvent(myClient *qqclient.MyClient) *Event {
	return &Event{myClient: myClient}
}

func (event *Event) Start() {
	event.subscribeChartEvent()
}

func (event *Event) subscribeEchoEvent() {
	myClient := event.myClient.Client
	myClient.SubscribeEventHandler(func(client *client.QQClient, event *message.GroupMessage) {
		msg := event.ToString()
		if strings.HasPrefix(msg, "/pjsk-echo") {
			log.Println("收到指令:" + msg)
			fields := strings.Fields(strings.TrimSpace(msg))
			echo := fields[1]
			_, err := myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText(echo)})
			if err != nil {
				if err != nil {
					log.Println("发送群消息异常", err)
				}
			}
		}
	})
}

func (event *Event) subscribeChartEvent() {
	myClient := event.myClient.Client
	myClient.SubscribeEventHandler(func(client *client.QQClient, event *message.GroupMessage) {
		msg := event.ToString()
		if strings.HasPrefix(msg, "/pjsk-chart") {
			log.Println("收到指令:" + msg)
			//解析指令
			fields := strings.Fields(strings.TrimSpace(msg))
			id, err := strconv.Atoi(fields[1])
			if err != nil {
				_, err := myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText("歌曲id无法正确解析")})
				if err != nil {
					log.Println("发送群消息异常", err)
				}
				return
			}
			level := fields[2]

			//发送请求
			params := url.Values{}
			params.Add("id", strconv.Itoa(id))
			params.Add("level", level)
			//发送请求，接收图片
			data, err := http.Get("http://localhost:9470/pjsk/charts" + "?" + params.Encode())
			if err != nil {
				_, err = myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText("请求失败，服务端没有正常响应")})
				if err != nil {
					log.Println("发送群消息异常", err)
				}
				return
			}

			//没有错误的话响应是图片，有错误的话是json
			contentType := data.Header.Get("Content-Type")
			if contentType == "image/png" {
				//读取图片字节流
				image, err := io.ReadAll(data.Body)
				if err != nil {
					log.Println(err)
					_, err := myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText("请求失败，服务端没有正确返回图片")})
					if err != nil {
						log.Println("发送群消息异常", err)
					}
					return
				}
				//发送图片
				_, err = myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewImage(image)})
				if err != nil {
					log.Println("发送群消息异常", err)
				}
			} else if strings.HasPrefix(contentType, "application/json") {
				//读取json
				json, err := io.ReadAll(data.Body)
				if err != nil {
					log.Println(err)
					return
				}
				_, err = myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText(string(json))})
				if err != nil {
					log.Println("发送群消息异常", err)
					return
				}
			} else {
				_, err = myClient.SendGroupMessage(event.GroupUin, []message.IMessageElement{message.NewText("发生错误，没有返回合理的响应")})
				if err != nil {
					log.Println(err)
				}
			}
		}
	})
}
