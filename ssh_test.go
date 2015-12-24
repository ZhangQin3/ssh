package ssh

import (
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	client, _ := Connect("10.89.255.1", "ethan", "pplab00")
	// client.SendCommand("pwd")
	// b, err := client.RecvUntil("~]", 3)

	b, err := client.Send("pwd", "~]", 3)

	fmt.Println(string(b), err)
}
