package dbConnect

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// BuildConnection establishs connection with mysql database and returns a connector.
func BuildConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:zpy19980412@tcp(localhost:3306)/EntryTask")
	if err != nil {
		panic(err.Error())
	}
	return db
}

// RetriveUserData returns 'userName','nickName','profilePicAddress' of a user giiven the userID.
func RetriveUserData(db *sql.DB, userID string) (string, string, string) {

	query := "select userName,nickName,profilePicAddress from users where id = " + userID + ""
	userName := ""
	nickName := ""
	profilePicAddress := ""
	err := db.QueryRow(query).Scan(&userName, &nickName, &profilePicAddress)
	if err != nil {
		panic(err)
	}
	return userName, nickName, profilePicAddress
}

// UpdateNickName updates user's nickname given a specific userID, returns an error if failed.
func UpdateNickName(db *sql.DB, userID string, nickName string) error {
	query := "update users set nickName = '" + nickName + "' where id = '" + userID + "'"
	// fmt.Println(query)

	update, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer update.Close()
	return nil
}

// UpdatePic updates user's profile picture address given a specific userID, returns an error if failed.
func UpdatePic(db *sql.DB, userID string, picAddress string) error {
	query := "update users set  profilePicAddress = '" + picAddress + "' where id = '" + userID + "'"
	// fmt.Println(query)

	update, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer update.Close()
	return nil
}

// VerifyIdentity checks if the given password matches the actual password of a user given the specific userID.
// If it matches, returns true,nil
// else, returns false,error message
func VerifyIdentity(db *sql.DB, userID string, pass string) (bool, error) {
	password := ""
	query := "select pass from users where id = '" + userID + "'"
	// fmt.Println(query)

	sel, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&password)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(tag.ID, tag.Pass)
	}

	if password == "" {
		// The case that the user doesn't exits. It doesn't have a password
		return false, fmt.Errorf("User does not exist")
	} else if password != pass {
		// The case that the password doesn't match
		return false, fmt.Errorf("Incorect password")
	} else {
		// Successful case
		return true, nil
	}

}
