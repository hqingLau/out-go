package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hqinglau/out-go/client"
	"github.com/hqinglau/out-go/server"
	"github.com/urfave/cli/v2"
)

func main() {
	// server的参数
	var (
		serverIPPort string // 服务器和客户端的通信端口
		dataPort     uint
	)

	// client的参数
	var (
		serverip         string
		serverport       uint
		dataport         uint
		clientLocalPort  uint // 客户端想要暴露的端口
		clientRemotePort uint // 本机映射至服务器的端口（需开放）
	)

	app := cli.NewApp()

	app.Usage = "内网穿透"
	app.Authors = []*cli.Author{{
		Name:  "hqinglau",
		Email: "hqinglau@gmail.com",
	}}

	app.Action = func(ctx *cli.Context) error {
		fmt.Println("请使用-h查看使用方法")
		return nil
	}

	app.Commands = []*cli.Command{
		{
			Name:  "server",
			Usage: "服务器端",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "port",
					Required:    true,
					Usage:       "server和client的通信端口",
					Destination: &serverIPPort,
				},

				&cli.UintFlag{
					Name:        "dataPort",
					Required:    true,
					Usage:       "server和client的数据端口",
					Destination: &dataPort,
				},
			},

			Action: func(ctx *cli.Context) error {
				s := server.Server{
					ServerIPPort: serverIPPort,
					DataPort:     dataPort,
				}
				s.Run()
				return nil
			},
		},
		{
			Name:  "client",
			Usage: "客户端",

			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:        "localPort",
					Required:    true,
					Usage:       "本地暴露端口",
					Destination: &clientLocalPort,
				},
				&cli.UintFlag{
					Name:        "remotePort",
					Required:    true,
					Usage:       "本地映射至远程的端口",
					Destination: &clientRemotePort,
				},
				&cli.UintFlag{
					Name:        "serverPort",
					Required:    true,
					Usage:       "server端和client的通信端口",
					Destination: &serverport,
				},
				&cli.UintFlag{
					Name:        "dataPort",
					Required:    true,
					Usage:       "server端和client的通信端口",
					Destination: &dataport,
				},
				&cli.StringFlag{
					Name:        "serverip",
					Required:    true,
					Usage:       "server ip",
					Destination: &serverip,
				},
			},

			Action: func(ctx *cli.Context) error {
				c := client.Client{
					ServerIP:       serverip,
					ServerDataPort: dataport,
					ServerPort:     serverport,
					LocalPort:      clientLocalPort,
					RemotePort:     clientRemotePort,
				}
				c.Run()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
