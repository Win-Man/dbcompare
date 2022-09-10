/*
 * Created: 2020-04-11 16:51:05
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
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

type Myssh struct {
	Host           string
	User           string
	Password       string
	Port           int
	Type           string
	KeyPath        string
	ConnectTimeout int64
	MyClient       *ssh.Client
	MySession      *ssh.Session
}

func (this *Myssh) ExecCommand(cmd string) (res []byte, err error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            this.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if this.Type == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(this.Password)}
	} else {
		log.Fatalf("Auth failed %s", this.Type)
	}

	// dial
	addr := fmt.Sprintf("%s:%d", this.Host, this.Port)
	this.MyClient, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("Create ssh client failed", err)
	}
	defer this.MyClient.Close()

	// create session
	this.MySession, err = this.MyClient.NewSession()
	if err != nil {
		log.Fatal("Create ssh session failed", err)
	}
	defer this.MySession.Close()

	// exec command
	res, err = this.MySession.CombinedOutput(cmd)
	if err != nil {
		log.Fatal("Exec command failed", err)
	}
	return res, err
}
