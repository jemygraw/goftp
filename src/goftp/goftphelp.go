package goftp

import (
	"fmt"
)

var FTP_CLIENT_HELP = map[string]string{
	FCC_HELP:          "print local help information",
	FCC_QUESTION_MARK: "print local help information",
}

type GoFtpClientHelp struct {
}

func (this *GoFtpClientHelp) version() {
	fmt.Println("GoFtpClient v1.0\r\n多科学堂出品\r\nhttps://github.com/jemygraw/goftp")
}

func (this *GoFtpClientHelp) help() {

}

func (this *GoFtpClientHelp) cmdHelp(cmdNames []string) {
	for _, cmdName := range cmdNames {
		if cmdHelpDoc, ok := FTP_CLIENT_HELP[cmdName]; ok {
			fmt.Println(cmdName, "\t", cmdHelpDoc)
		} else {
			fmt.Println(cmdName, "\tN/A")
		}
	}
}
