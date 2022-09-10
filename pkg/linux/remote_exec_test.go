/*
 * Created: 2020-04-11 16:53:49
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package linux

import (
	"fmt"
	"testing"
)

func TestExecCommand(t *testing.T) {
	// var tests = []struct {
	// 	x    int
	// 	want bool
	// }{
	// 	{121, true},
	// 	{-121, false},
	// 	{10, false},
	// 	{1, true},
	// 	{100, false},
	// }

	// for _, tt := range tests {
	// 	testname := fmt.Sprintf("%d", tt.x)
	// 	t.Run(testname, func(t *testing.T) {
	// 		ans := isPalindrome(tt.x)
	// 		if ans != tt.want {
	// 			t.Errorf("isPalindrome(%d) got %t,want %t", tt.x, ans, tt.want)
	// 		}
	// 	})

	// }
	// Host           string
	// User           string
	// Password       string
	// Port           int
	// Type           string
	// KeyPath        string
	// ConnectTimeout int64
	// MyClient *ssh.Client
	// MySession *ssh.Session
	t.Run("TestExecCommand", func(t *testing.T) {
		myssh := Myssh{"172.16.4.69", "root", "pingcap!@#", 22, "password", "", 0, nil, nil}
		res, err := myssh.ExecCommand("cat /home/tidb/config")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(res))
	})
}
