package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"pengyuzhao/Entrytask/Tcp/dbConnect"

	"google.golang.org/protobuf/proto"
)

func intToByteArray(a uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, a)
	return bs
}

func readPackages(c net.Conn) *Command {
	comm := &Command{}
	dataLenBytes := make([]byte, 4)
	_, err1 := io.ReadFull(c, dataLenBytes)
	if err1 != nil {
		panic(err1)
	}
	dataLen := binary.LittleEndian.Uint32(dataLenBytes)
	fmt.Println("received package length is :", dataLen)
	dataBytes := make([]byte, dataLen)
	_, err2 := io.ReadFull(c, dataBytes)
	if err2 != nil {
		panic(err2)
	}

	err3 := proto.Unmarshal(dataBytes, comm)
	if err3 != nil {
		panic(err3)
	}
	return comm
}
func sendPackages(c net.Conn, res *Response) {
	data, err := proto.Marshal(res)
	if err != nil {
		panic(err)
	}
	fmt.Println("size of data sending is :", len(data))
	dataLen := intToByteArray(uint32(len(data)))
	c.Write(dataLen)
	c.Write(data)
}
func preProcessingCommands(comm *Command) *Response {
	var res *Response
	switch comm.GetCommandType() {
	case Command_VERIFYUSER:
		res = handleLoginRequest(comm.GetUser())
	case Command_RETREIVEUSERINFO:
		res = handleRetriveUserInfoRequest(comm.GetUser())
	case Command_UPDATEPIC:
		res = handleUpdateUserPic(comm.GetUser())
	case Command_UPDATENICKNAME:
		res = handleUpdateUserNickName(comm.GetUser())
	}
	return res
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		comm := readPackages(c)
		fmt.Println("type of command is :", comm.GetCommandType())
		response := preProcessingCommands(comm)
		sendPackages(c, response)
		break
	}
	c.Close()
}
func handleRetriveUserInfoRequest(user *User) *Response {
	dbConn := dbConnect.BuildConnection()
	userName, nickName, profilePicAddress := dbConnect.RetriveUserData(dbConn, user.GetID())
	res := &Response{
		ResponseType: Response_SUCCESSFULL,
		User: &User{
			ID:             user.GetID(),
			UserName:       userName,
			NickName:       nickName,
			PictureAddress: profilePicAddress,
		},
	}
	return res
}

func handleUpdateUserNickName(user *User) *Response {
	res := &Response{
		ResponseType: Response_SUCCESSFULL,
	}
	dbConn := dbConnect.BuildConnection()
	err := dbConnect.UpdateNickName(dbConn, user.GetID(), user.NickName)
	if err != nil {
		fmt.Println(err.Error())
		res.ResponseType = Response_INVALIDUSERNAME
	}
	return res

}

func handleUpdateUserPic(user *User) *Response {
	res := &Response{
		ResponseType: Response_SUCCESSFULL,
	}
	dbConn := dbConnect.BuildConnection()
	err := dbConnect.UpdatePic(dbConn, user.GetID(), user.NickName)
	if err != nil {
		fmt.Println(err.Error())
		res.ResponseType = Response_INVALIDUSERNAME
	}
	return res
}

func handleLoginRequest(user *User) *Response {
	dbConn := dbConnect.BuildConnection()
	isValid, err := dbConnect.VerifyIdentity(dbConn, user.GetID(), user.GetPassword())
	res := &Response{}
	if isValid {
		res.ResponseType = Response_SUCCESSFULL
	} else if err.Error() == "Incorect password" {
		res.ResponseType = Response_INVALIDPASSWORD
	} else if err.Error() == "User does not exist" {
		res.ResponseType = Response_INVALIDUSERNAME
	}
	return res
}

func main() {
	l, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
