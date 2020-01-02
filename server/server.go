package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"

	pb "gRPC_stream/pb"

	"google.golang.org/grpc"
)

var musics = map[string]pb.MusicInfo{
	"lily": {
		MusicName: "lily",
		MusicType: "foreign",
		MusicUrl:  "https://www.youtube.com/results?search_query=lily",
	},
}

//ReturnType: Default 0, In list 1, Not in list 2, list the list 3, save the list 4, stop server 5, error 6
var album = &pb.MusicResponse{
	MusicList: []*pb.MusicInfo{&pb.MusicInfo{
		MusicName: "lily",
		MusicType: "foreing",
		MusicUrl:  "https://www.youtube.com/results?search_query=" + "lily",
	}},
	ReturnType: 0,
}

type Server struct {
}

func Find(s []string, substr string) (int, bool) {
	for i, v := range s {
		if substr == v {
			return i, true
		}
	}
	return -1, false

}

// 之前提到Go只要有完成interface的方法, 就等於繼承了該接口
// GetUserInfo(context.Context, *UserRequest) (*UserResponse, error)

func (s *Server) GetMusicInfo(srv pb.MusicService_GetMusicInfoServer) (err error) {
	commands := []string{"list", "save", "exit"}

	for {

		in, err := srv.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("fail to recv: %v", err)
			return err
		}

		if strings.HasPrefix(in.MusicName, ";;") {
			_, found := Find(commands, strings.TrimLeft(in.MusicName, ";;"))
			if !found {
				album.ReturnType = 6
				album.ReturnMessage = "The command " + in.MusicName + " is not exsited."
				srv.Send(album)
				continue
			}
		}

		switch in.MusicName {
		case ";;exit":
			album.ReturnType = 5
			album.ReturnMessage = "music client leave."
			srv.Send(album)
			break
		case ";;list":
			album.ReturnType = 3
			album.ReturnMessage = "Music in Album:"
			srv.Send(album)
		case ";;save":
			var saveList []string
			for _, music := range album.MusicList {
				saveList = append(saveList, music.MusicName)
			}
			saveFile := strings.Join(saveList, "\n")
			err := ioutil.WriteFile("musicList.txt", []byte(saveFile), 0777)
			if err != nil {
				log.Printf("fail to save: %v", err)
			}
			album.ReturnType = 4
			album.ReturnMessage = "The musicList has saved."
			srv.Send(album)
		default:
			if _, ok := musics[in.MusicName]; !ok {
				musics[in.MusicName] = pb.MusicInfo{
					MusicName: in.MusicName,
					MusicType: "foreign",
					MusicUrl:  "https://www.youtube.com/results?search_query=" + in.MusicName,
				}
				music := musics[in.MusicName]
				album.MusicList = append(album.MusicList, &music)
				album.ReturnType = 2
				album.ReturnMessage = "The music " + in.MusicName + " has add to album."
				srv.Send(album)
			} else {
				album.ReturnType = 1
				album.ReturnMessage = "The music  " + in.MusicName + " has in album."
				srv.Send(album)
			}
		}

	}
	return nil
}

func main() {
	// 建構一個gRPC服務端實例
	grpcServer := grpc.NewServer()

	// 註冊服務
	pb.RegisterMusicServiceServer(grpcServer, &Server{})

	// 註冊端口來提供gRPC服務
	listen, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server is running.")
	grpcServer.Serve(listen)

}
