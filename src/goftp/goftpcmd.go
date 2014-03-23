package goftp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	FC_REQUEST_SUFFIX  string = "\r\n"
	FC_RESPONSE_BUFFER int    = 1024
)

//定义与ftp服务器进行交互的命令，前缀FC表示Ftp Command
const (
	FC_USER string = "USER" //USER login_name
	FC_PASS string = "PASS" //PASS login_pass
	FC_QUIT string = "QUIT" //QUIT
	FC_PWD  string = "PWD"  //PWD
	FC_CWD  string = "CWD"  //CWD remote_dir
	FC_LIST string = "LIST" //LIST remote_dir
	FC_PASV string = "PASV" //PASV
)

type GoFtpClientCmd struct {
	Name      string
	Params    []string
	Connected bool

	DefaultLocalWorkDir string
	LocalWorkDir        string
	Username            string

	FtpConn net.Conn

	GoFtpClientHelp
}

func (this *GoFtpClientCmd) welcome() {
	var data []byte = make([]byte, 1024)
	readCount, err := this.FtpConn.Read(data)
	if err == nil {
		fmt.Print(string(data[:readCount]))
		this.user()
		this.pass()
	} else {
		fmt.Println(err)
	}
}

func (this *GoFtpClientCmd) sendCmdRequest(ftpParams []string) {
	if this.Connected {
		var sendData = fmt.Sprint(strings.Join(ftpParams, " "), FC_REQUEST_SUFFIX)
		this.FtpConn.Write([]byte(sendData))
	} else {
		fmt.Println("Not connected.")
	}
}

func (this *GoFtpClientCmd) recvCmdResponse() (recvData string) {
	if this.Connected {
		var recvBytes []byte = make([]byte, FC_RESPONSE_BUFFER)
		readCount, err := this.FtpConn.Read(recvBytes)
		if err == nil {
			recvData = string(recvBytes[:readCount])
			fmt.Print(string(recvData))
		} else {
			fmt.Println(err)
		}
	}
	return
}

func (this *GoFtpClientCmd) user() {
	var remoteAddr = this.FtpConn.RemoteAddr().String()
	var portIndex = strings.LastIndex(remoteAddr, ":")
	fmt.Printf("Name (%s:%s):", remoteAddr[:portIndex], this.Username)
	var username string
	fmt.Scanln(&username)
	this.sendCmdRequest([]string{FC_USER, username})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) pass() {
	fmt.Print("Password:")
	var password string
	fmt.Scanln(&password)
	this.sendCmdRequest([]string{FC_PASS, password})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) lcd() {
	var paramCount = len(this.Params)
	if paramCount == 0 || paramCount == 1 {
		if paramCount == 0 {
			this.LocalWorkDir = this.DefaultLocalWorkDir
			fmt.Println("Local directory now:", this.LocalWorkDir)
		} else {
			var path = this.Params[0]
			if !filepath.IsAbs(path) {
				path = filepath.Join(this.DefaultLocalWorkDir, path)
			}
			fiInfo, err := os.Stat(path)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				if fiInfo.IsDir() {
					this.LocalWorkDir = path
					fmt.Println("Local directory now:", path)
				} else {
					fmt.Println("ftp: Can't chdir `", path, "': No such file or directory")
				}
			}
		}
	} else {
		this.cmdUsage(this.Name)
	}
}

func (this *GoFtpClientCmd) pwd() {
	this.sendCmdRequest([]string{FC_PWD})
	this.recvCmdResponse()
}

func (this *GoFtpClientCmd) cwd() {
	var paramCount = len(this.Params)
	if paramCount == 0 {
		fmt.Println("(remote-directory)")
		var remoteDir string
		fmt.Scanln(&remoteDir)
		if remoteDir != "" {
			this.sendCmdRequest([]string{FC_CWD, remoteDir})
			this.recvCmdResponse()
		} else {
			this.cmdUsage(this.Name)
		}
	} else if paramCount > 1 {
		this.cmdUsage(this.Name)
	} else {
		this.sendCmdRequest([]string{FC_CWD, this.Params[0]})
		this.recvCmdResponse()
	}
}

func (this *GoFtpClientCmd) ls() {
	var paramCount = len(this.Params)
	if paramCount >= 0 && paramCount <= 2 {
		var resultOutputFile string
		var remoteDir string
		if paramCount == 1 {
			remoteDir = this.Params[0]
		} else if paramCount == 2 {
			remoteDir = this.Params[0]
			resultOutputFile = this.Params[1]
		}
		var outputFile *os.File
		var err error
		if resultOutputFile != "" {
			if !filepath.IsAbs(resultOutputFile) {
				resultOutputFile = filepath.Join(this.LocalWorkDir, resultOutputFile)
			}
			outputFile, err = os.Create(resultOutputFile)
			if err != nil {
				fmt.Println("ftp: Can't access `", resultOutputFile, "': No such file or directory")
			}
		}

		if err == nil {
			pasvHost, pasvPort := this.pasv()
			if pasvHost != "" {
				this.sendCmdRequest([]string{FC_LIST, remoteDir})
				this.recvCmdResponse()
				var pasvRespData = this.getPasvData(pasvHost, pasvPort)
				if outputFile != nil {
					var bWriter = bufio.NewWriter(outputFile)
					bWriter.WriteString(string(pasvRespData))
					bWriter.Flush()
					outputFile.Close()
				} else {
					fmt.Print(string(pasvRespData))
				}
				this.recvCmdResponse()
			}
		}
	} else {
		this.cmdUsage(this.Name)
	}
}

func (this *GoFtpClientCmd) pasv() (pasvHost string, pasvPort int) {
	if this.Connected {
		this.sendCmdRequest([]string{FC_PASV})
		var recvData = this.recvCmdResponse()
		var startIndex = strings.Index(recvData, "(")
		var endIndex = strings.LastIndex(recvData, ")")
		var pasvDataStr = recvData[startIndex+1 : endIndex]
		var pasvDataParts = strings.Split(pasvDataStr, ",")
		pasvHost = strings.Join(pasvDataParts[:4], ".")
		var p1, _ = strconv.Atoi(pasvDataParts[4])
		var p2, _ = strconv.Atoi(pasvDataParts[5])
		pasvPort = p1*256 + p2
	} else {
		fmt.Println("Not connected.")
	}
	return
}

func (this *GoFtpClientCmd) getPasvData(pasvHost string, pasvPort int) (pasvRespData []byte) {
	pasvConn, pasvConnErr := net.DialTimeout("tcp", net.JoinHostPort(pasvHost, strconv.Itoa(pasvPort)),
		time.Duration(DIAL_FTP_SERVER_TIMEOUT_SECONDS)*time.Second)
	if pasvConnErr != nil {
		fmt.Println(pasvConnErr.Error())
	} else {
		var bReader = bufio.NewReader(pasvConn)
		pasvRespData = make([]byte, FC_RESPONSE_BUFFER)
		for {
			line, err := bReader.ReadString('\n')
			pasvRespData = append(pasvRespData, []byte(line)...)
			if err == io.EOF {
				break
			}
		}
		pasvConn.Close()
	}
	return
}

func (this *GoFtpClientCmd) disconnect() {
	this.close()
}

func (this *GoFtpClientCmd) close() {
	if this.FtpConn != nil {
		this.sendCmdRequest([]string{FC_QUIT})
		this.recvCmdResponse()

		this.FtpConn = nil
		this.Name = ""
		this.Params = nil
		this.Connected = false
	}
}
