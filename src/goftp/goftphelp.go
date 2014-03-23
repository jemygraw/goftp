package goftp

import (
	"fmt"
)

type GoFtpClientHelp struct {
}

func (this *GoFtpClientHelp) version() {
	fmt.Println("GoFtpClient v1.0\r\n多科学堂出品\r\nhttps://github.com/jemygraw/goftp")
}

func (this *GoFtpClientHelp) help() {

}

func (this *GoFtpClientHelp) cmdHelp(cmdName string) {

}
