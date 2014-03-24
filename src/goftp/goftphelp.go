package goftp

import (
	"fmt"
	"strings"
)

//ftp客户端命令的帮助信息
var FTP_CLIENT_CMD_HELP = map[string]string{
	FCC_HELP:          "print local help information",
	FCC_QUESTION_MARK: "print local help information",
	FCC_CD:            "change remote working directory",
	FCC_LS:            "list contents of remote path",
	FCC_LCD:           "change local working directory",
	FCC_OPEN:          "connect to remote ftp server",
	FCC_USER:          "send new user information",
	FCC_USAGE:         "show usage of ftp command",
}

//ftp客户端命令的使用方法
//其中带`[]`的参数都是可选参数
var FTP_CLIENT_CMD_USAGE = map[string]string{
	FCC_HELP:          "help [cmd1],[cmd2],...",
	FCC_QUESTION_MARK: "? [cmd1],[cmd2],...",
	FCC_CD:            "cd remote_dir",
	FCC_LS:            "ls [remote_dir|remote_file] [local_output_file]",
	FCC_LCD:           "lcd [local_directory]",
	FCC_OPEN:          "open remote_host [port]",
	FCC_USER:          "user username [password] [account]",
}

type GoFtpClientHelp struct {
}

func (this *GoFtpClientHelp) version() {
	fmt.Println("GoFtpClient v1.0\r\n多科学堂出品\r\nhttps://github.com/jemygraw/goftp")
}

func (this *GoFtpClientHelp) help() {

}

func (this *GoFtpClientHelp) cmdHelp(cmdNames ...string) {
	for _, cmdName := range cmdNames {
		cmdName = strings.ToLower(cmdName)
		if cmdHelpDoc, ok := FTP_CLIENT_CMD_HELP[cmdName]; ok {
			fmt.Println(cmdName, "\t", cmdHelpDoc)
		} else {
			fmt.Println("?Invalid help command `", cmdName, "'")
		}
	}
}

func (this *GoFtpClientHelp) cmdUsage(cmdNames ...string) {
	for _, cmdName := range cmdNames {
		cmdName = strings.ToLower(cmdName)
		if cmdUsageDoc, ok := FTP_CLIENT_CMD_USAGE[cmdName]; ok {
			fmt.Println("Usage:", cmdUsageDoc)
		} else {
			fmt.Println("?Invalid usage command `", cmdName, "'")
		}
	}
}
