package main

import (
	"fmt"
	pb "gRPC_stream/pb"
	"gRPC_stream/server/youtube"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"

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
	HelpCommands
	RemoveMusic
)

type Server struct {
	lock sync.Mutex
}

func FindSubString(s []string, substr string) (int, bool) {
	for i, v := range s {
		if substr == v {
			return i, true
		}
	}
	return -1, false

}

func readMusicList() (*pb.MusicResponse, map[string]pb.MusicInfo) {
	var musicList []*pb.MusicInfo
	var musics = map[string]pb.MusicInfo{
		"default": {
			MusicName: "MusicName",
			MusicType: "MusicType",
			MusicUrl:  "https://www.youtube.com/results?search_query=MusicUrl",
		},
	}

	readFile, err := ioutil.ReadFile("musicList.txt")
	if err != nil {
		log.Printf("fail to read file: %v", err)
	}

	rFileLine := strings.Split(string(readFile), "\n")
	for _, rFileSpace := range rFileLine {
		if rFileSpace != "" {
			name := strings.TrimRight(strings.Split(rFileSpace, "https")[0], " ")
			musicInfo := pb.MusicInfo{
				MusicName: name,
				MusicType: "foreign",
				MusicUrl:  "https" + strings.Split(rFileSpace, "https")[1],
			}
			musics[name] = musicInfo
			musicList = append(musicList, &musicInfo)
		}
	}
	musicResponse := pb.MusicResponse{
		MusicList:     musicList,
		ReturnType:    Default,
		ReturnMessage: "",
	}

	return &musicResponse, musics

}

func saveMusicList(musicList []*pb.MusicInfo) {
	var saveList []string
	for _, music := range musicList {
		saveList = append(saveList, music.MusicName+" "+music.MusicUrl)
	}
	saveFile := strings.Join(saveList, "\n")
	err := ioutil.WriteFile("musicList.txt", []byte(saveFile), 0777)
	if err != nil {
		log.Printf("fail to save file: %v", err)
	}
}

func setDefaultMusicList(album *pb.MusicResponse, musics map[string]pb.MusicInfo) (*pb.MusicResponse, map[string]pb.MusicInfo) {
	musicDefault := &pb.MusicInfo{
		MusicName: "MusicName",
		MusicType: "MusicType",
		MusicUrl:  "https://www.youtube.com/results?search_query=MusicUrl",
	}
	musics = map[string]pb.MusicInfo{
		"default": {
			MusicName: musicDefault.MusicName,
			MusicType: musicDefault.MusicType,
			MusicUrl:  musicDefault.MusicUrl,
		},
	}
	var musicList []*pb.MusicInfo
	musicList = append(musicList, musicDefault)

	album = &pb.MusicResponse{
		MusicList:     musicList,
		ReturnType:    Default,
		ReturnMessage: "",
	}
	return album, musics
}

/*
func removeMusic(album *pb.MusicResponse,musics map[string]pb.MusicInfo,s string)(*pb.MusicResponse,map[string]pb.MusicInfo){
	delete(musics,s)
	album.MusicList


	return
}*/

// 之前提到Go只要有完成interface的方法, 就等於繼承了該接口
// GetUserInfo(context.Context, *UserRequest) (*UserResponse, error)

func (s *Server) GetMusicInfo(srv pb.MusicService_GetMusicInfoServer) (err error) {
	commands := []string{"list", "save", "exit", "clear", "help", "remove"}

	album, musics := readMusicList()

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
			_, found := FindSubString(commands, strings.TrimLeft(in.MusicName, ";;"))
			if !found {
				album.ReturnType = Error
				album.ReturnMessage = "The command " + in.MusicName + " is not exsited."
				srv.Send(album)
				continue
			}
		}

		switch in.MusicName {
		case ";;exit":
			album.ReturnType = StopServer
			album.ReturnMessage = "music client leave."
			srv.Send(album)
			return err
		case ";;list":
			album.ReturnType = ListList
			album.ReturnMessage = "Music in Album:"
			srv.Send(album)
		case ";;save":
			s.lock.Lock()
			saveMusicList(album.MusicList)
			s.lock.Unlock()

			album.ReturnType = SaveList
			album.ReturnMessage = "The musicList is saved."
			srv.Send(album)
		case ";;clear":
			album, musics = setDefaultMusicList(album, musics)
			album.ReturnType = ClearList
			album.ReturnMessage = "The musicList is cleared."
			srv.Send(album)
		case ";;help":
			album.ReturnType = HelpCommands
			album.ReturnMessage = strings.Join(commands, ",")
			srv.Send(album)
		case ";;remove":
			continue

		default:
			id, name := youtube.SearchVideo(in.MusicName, 1)
			if _, ok := musics[name]; !ok {
				s.lock.Lock()
				musics[name] = pb.MusicInfo{
					MusicName: name,
					MusicType: "foreign",
					MusicUrl:  "https://www.youtube.com/watch?v=" + id,
				}
				music := musics[name]
				album.MusicList = append(album.MusicList, &music)
				s.lock.Unlock()
				album.ReturnType = NotInList
				album.ReturnMessage = "The music " + name + " is add to album."
				srv.Send(album)
			} else {
				album.ReturnType = InList
				album.ReturnMessage = "The music  " + name + " has in album."
				srv.Send(album)
			}
		}

	}
}

func main() {

	// 建構一個gRPC服務端實例
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
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
