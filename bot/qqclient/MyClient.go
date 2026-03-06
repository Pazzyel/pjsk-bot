package qqclient

import (
	"log"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/auth"
)

type MyClient struct {
	Client *client.QQClient
}

func NewMyClient(qqNumber uint32, password string) *MyClient {
	qqclient := client.NewClient(qqNumber, password)
	qqclient.AddSignServer("https://sign.lagrangecore.org/api/sign")
	qqclient.AddSignServer("https://sign.lagrangecore.org/api/sign/25765")
	qqclient.AddSignServer("https://sign.0w0.ing/api/sign")
	qqclient.AddSignServer("https://sign.0w0.ing/api/sign/25765")
	qqclient.UseVersion(auth.AppList["linux"]["3.1.2-13107"])
	qqclient.UseDevice(auth.NewDeviceInfo(114514130))
	myClient := &MyClient{Client: qqclient}
	myClient.LoginMyClient()
	return myClient
}

func (this *MyClient) LoginMyClient() {
	_, err := this.Client.PasswordLogin()
	if err != nil {
		log.Println("使用密码登录失败，", err)
	}

}
