package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"pengyuzhao/Entrytask/Tcp/dbConnect"

	"google.golang.org/protobuf/proto"
)

// Converts int to byte array.
func intToByteArray(a uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, a)
	return bs
}

// Read Command package sent from Http Server.
// Returns a @Commmand object
func readPackages(c net.Conn) *Command {
	comm := &Command{}
	//Read first package sent from Http server, which shows the length of command package
	dataLenBytes := make([]byte, 4)
	_, err1 := io.ReadFull(c, dataLenBytes)
	if err1 != nil {
		panic(err1)
	}
	dataLen := binary.LittleEndian.Uint32(dataLenBytes)
	// fmt.Println("received package length is :", dataLen)

	//Read the actual command package
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

// Send Response package to Http Server.
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

// Identify type of command, then distribute to handling fucntion
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

//Handle Tcp connection from Http server
//Receive Command package
//Processing the Command
//Send Response to Http Server
func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		comm := readPackages(c)
		fmt.Println("type of command is :", comm.GetCommandType())
		response := preProcessingCommands(comm)
		sendPackages(c, response)
	}
}

//Handle RetriveUserInfo Command, which should return a response containing
//user's userName,nickName and profile picture address of the given userID
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

//Handle UpdateUserNick Command. It should return  a response showing the update is
//successful or fialed with error message
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

//Handle UpdateProfilePic Command. It should return  a response showing the update is
//successful or fialed with error message
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

//Handle verifyUser Command. It should verify the user's identity.
//If the given user is not a valid user, it should return the response
//containing corresbonding response type.
func handleLoginRequest(user *User) *Response {
	dbConn := dbConnect.BuildConnection()
	isValid, err := dbConnect.VerifyIdentity(dbConn, user.GetID(), user.GetPassword())
	res := &Response{}
	if isValid {
		//The given user is valid
		res.ResponseType = Response_SUCCESSFULL
	} else if err.Error() == "Incorect password" {
		//The given password is incorrect
		res.ResponseType = Response_INVALIDPASSWORD
	} else if err.Error() == "User does not exist" {
		//The given userId is incorrect, the user doesn't exist in the
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
