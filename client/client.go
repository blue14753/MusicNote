package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	pb "gRPC_stream/pb"

	"google.golang.org/grpc"
)

const (
	Default int32 = iota
	InList
	NotInList
	ListList
	SaveList
	StopServer
	Error
	ClearList
)

func GetMusicInfo(c pb.MusicServiceClient) {
	var err error
	stream, err := c.GetMusicInfo(context.Background())
	in := bufio.NewReader(os.Stdin)

	if err != nil {
		log.Printf("fail to call: %v", err)
		return
	}

	for {

		nameInput, err := in.ReadString('\n')
		if err != nil {
			log.Printf("failed to  read: %v", err)
		}
		nameInput = strings.TrimRight(nameInput, "\n")

		stream.Send(&pb.MusicInfo{MusicName: nameInput})
		if err != nil {
			log.Printf("failed to send: %v", err)
			break
		}

		reply, err := stream.Recv()
		if err != nil {
			log.Printf("fail to recv: %v", err)
			break
		}

		switch reply.ReturnType {
		case InList, NotInList, SaveList, Error, ClearList:
			fmt.Println(reply.ReturnMessage)
		case ListList:
			fmt.Println(reply.ReturnMessage)
			for _, music := range reply.MusicList {
				fmt.Println(music.MusicName + " " + music.MusicUrl)
			}
		case StopServer:
			fmt.Println(reply.ReturnMessage)
			return
		}
	}

}

func main() {
	// 透過Dial()負責跟gRPC服務端建立起連線
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// 注入連線, 返回UserServiceClient對象
	client := pb.NewMusicServiceClient(conn)
	// 接著就能像一般調用方法那樣呼叫了

	fmt.Println("Please input music name:")

	GetMusicInfo(client)
}
