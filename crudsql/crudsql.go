package crudsql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"os"
	"time"
)

type messageRow struct {
	MessageID uint      `json:"messageID"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userID"`
	Entry     string    `json:"entry"`
}

type Crudsql struct {
	Con     pgx.Conn
	Msg     [10]messageRow // While this can be returned after ReadMsg(), explicitly getting it later leads to slightly more readable code
	MsgJSON []byte
}

func (c *Crudsql) CreateMsg(userID string, entry string) {
	commandTag, err := c.Con.Exec(context.Background(), fmt.Sprintf(
		"INSERT INTO Messages ("+
			"userID, "+
			"entry) "+
			"VALUES ("+
			"'%1s', "+
			"'%2s')", userID, entry))
	if err != nil {
		fmt.Println(commandTag)
		panic(err)
	}
	if commandTag.RowsAffected() != 1 {
		panic("Inserted too many rows!")
	}
}
func (c *Crudsql) ReadMsg() {
	rows, err := c.Con.Query(context.Background(),
		"SELECT m.messageID, "+
			"m.timestamp, "+
			"m.userID, "+
			"m.entry "+
			"FROM Messages AS m "+
			"ORDER BY m.timestamp DESC "+
			"LIMIT 10")

	// We can either iterate '(i=0,i<10,i++)' or 'for row.Next'
	// Former enforces a solid limit of ten rows
	// Latter enforces every row returned is parsed
	// Here the latter is used as the limit of ten rows is already enforced in the SQL query
	// Either should be acceptable, as the bottleneck on a scaled system is generally in the SQL execution

	i := 0

	for rows.Next() {
		err = rows.Scan(&c.Msg[i].MessageID, &c.Msg[i].Timestamp, &c.Msg[i].UserID, &c.Msg[i].Entry)
		i++
	}
	a, _ := json.Marshal(c.Msg)
	c.MsgJSON = a
	fmt.Println(string(c.MsgJSON))
	defer rows.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query Failed: %v\n", err)
		os.Exit(1)
	}

}
func (c *Crudsql) UpdateMsg(messageID uint, userID string, entry string) {
	commandTag, err := c.Con.Exec(context.Background(), fmt.Sprintf(
		"UPDATE Messages "+
			"SET "+
			"userID='%1s', "+
			"entry='%2s' "+
			"WHERE "+
			"messageID=%3d", userID, entry, messageID))
	if err != nil {
		panic(err)
	}
	if commandTag.RowsAffected() != 1 {
		panic("No row to update!")
	}
}
func (c *Crudsql) DeleteMsg(messageID uint) string {
	commandTag, err := c.Con.Exec(context.Background(), fmt.Sprintf(
		"DELETE FROM Messages "+
			"WHERE messageID=%1d", messageID))
	if err != nil {
		panic(err)
	}
	if commandTag.RowsAffected() != 1 {
		return ("No row to delete!")
	}
	return ("Row deleted")
}
func (c *Crudsql) GetConnection(username string, password string, dbname string) {
	conString := fmt.Sprintf("user=%1s password=%2s dbname=%3s sslmode=disable", username, password, dbname) // for local development

	connect, err := pgx.Connect(context.Background(), conString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	c.Con = *connect
}
