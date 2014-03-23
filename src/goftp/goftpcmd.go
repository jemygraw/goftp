package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//ftp服务器默认监听端口号
const (
	FTP_SERVER_DEFAULT_LISTENING_PORT = 21
)

const (
	FC_REQUEST_SUFFIX  string = "\r\n"
	FC_RESPONSE_BUFFER int    = 1024
)

const (
	FC_RESP_CODE_ENTER_PASSIVE_MODE int = 227
)

//定义与ftp服务器进行交互的命令，前缀FC表示Ftp Command
const (
	FC_USER string = "USER" //USER login_name
	FC_PASS string = "PASS" //PASS login_pass
	FC_ACCT string = "ACCT" //ACCT account_name
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
		//提示输入登录名
		var remoteAddr = this.FtpConn.RemoteAddr().String()
		var portIndex = strings.LastIndex(remoteAddr, ":")
		fmt.Printf("Name (%s:%s):", remoteAddr[:portIndex], this.Username)
		var username string
		fmt.Scanln(&username)
		this.sendCmdRequest([]string{FC_USER, username})
		this.recvCmdResponse()
		//提示输入登录密码
		fmt.Print("Password:")
		var password string
		fmt.Scanln(&password)
		this.sendCmdRequest([]string{FC_PASS, password})
		this.recvCmdResponse()
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
			fmt.Println("ftp:", err)
		}
	}
	return
}

func (this *GoFtpClientCmd) open() {
	if this.Connected {
		var remoteAddr = this.FtpConn.RemoteAddr().String()
		var portIndex = strings.LastIndex(remoteAddr, ":")
		fmt.Println("Already connected to ", remoteAddr[:portIndex], ", use close first.")
	} else {
		var paramCount = len(this.Params)
		var ftpHost string
		var ftpPort int
		if paramCount == 0 {
			fmt.Print("(To) ")
			cmdReader := bufio.NewReader(os.Stdin)
			cmdStr, err := cmdReader.ReadString('\n')
			cmdStr = strings.Trim(cmdStr, "\r\n")
			if err == nil && cmdStr != "" {
				cmdParts := strings.Fields(cmdStr)
				cmdPartCount := len(cmdParts)
				if cmdPartCount == 1 {
					ftpHost = cmdParts[0]
					ftpPort = FTP_SERVER_DEFAULT_LISTENING_PORT
				} else if cmdPartCount == 2 {
					ftpHost = cmdParts[0]
					port, err := strconv.Atoi(cmdParts[1])
					if err != nil {
						this.cmdUsage(this.Name)
					} else {
						ftpPort = port
					}
				}
			} else {
				this.cmdUsage(this.Name)
			}
		} else if paramCount == 1 {
			ftpHost = this.Params[0]
			ftpPort = FTP_SERVER_DEFAULT_LISTENING_PORT
		} else if paramCount == 2 {
			ftpHost = this.Params[0]
			port, err := strconv.Atoi(this.Params[1])
			if err != nil {
				this.cmdUsage(this.Name)
			} else {
				ftpPort = port
			}
		}

		//建立ftp连接
		if ftpHost != "" {
			ips, lookupErr := net.LookupIP(ftpHost)
			if lookupErr != nil {
				fmt.Println("ftp: Can't lookup host `", ftpHost, "'")
			} else {
				var port = strconv.Itoa(ftpPort)
				for _, ip := range ips {
					conn, connErr := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port),
						time.Duration(DIAL_FTP_SERVER_TIMEOUT_SECONDS)*time.Second)
					if connErr != nil {
						fmt.Println("Trying ", ip, "...")
						fmt.Println("ftp:", connErr.Error())
					} else {
						fmt.Println("Connected to", ip, ".")
						var sysUser, _ = user.Current()
						this.FtpConn = conn
						this.Connected = true
						this.Username = sysUser.Username
						this.DefaultLocalWorkDir = sysUser.HomeDir
						this.LocalWorkDir = sysUser.HomeDir

						this.welcome()
						break
					}
				}
			}
		}
	}
}

func (this *GoFtpClientCmd) parseCmdResponse(respData string) (ftpRespCode int, err error) {
	var recvDataParts = strings.Fields(respData)
	if len(recvDataParts) > 0 {
		ftpRespCode, err = strconv.Atoi(recvDataParts[0])
	}
	return
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
				fmt.Println("ftp:", err.Error())
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

func (this *GoFtpClientCmd) user() {
	var paramCount = len(this.Params)
	var username string
	var password string
	var account string
	if paramCount == 0 {
		fmt.Print("Username:")
		fmt.Scanln(&username)
		username = strings.Trim(username, "\r\n")
		if username == "" {
			this.cmdUsage(this.Name)
		} else {
			this.sendCmdRequest([]string{FC_USER, username})
			this.recvCmdResponse()
			fmt.Print("Password:")
			fmt.Scanln(&password)
			this.sendCmdRequest([]string{FC_PASS, password})
			this.recvCmdResponse()
		}
	} else if paramCount == 1 {
		username = this.Params[0]
		this.sendCmdRequest([]string{FC_USER, username})
		this.recvCmdResponse()
		fmt.Print("Password:")
		fmt.Scanln(&password)
		this.sendCmdRequest([]string{FC_PASS, password})
		this.recvCmdResponse()
	} else if paramCount == 2 {
		username = this.Params[0]
		password = this.Params[1]
		this.sendCmdRequest([]string{FC_USER, username})
		this.recvCmdResponse()
		this.sendCmdRequest([]string{FC_PASS, password})
		this.recvCmdResponse()
	} else if paramCount == 3 {
		username = this.Params[0]
		password = this.Params[1]
		account = this.Params[2]
		this.sendCmdRequest([]string{FC_USER, username})
		this.recvCmdResponse()
		this.sendCmdRequest([]string{FC_PASS, password})
		this.recvCmdResponse()
		this.sendCmdRequest([]string{FC_ACCT, account})
		this.recvCmdResponse()
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
			pasvHost, pasvPort, ftpRespCode, errPasv := this.pasv()
			if errPasv == nil {
				if pasvHost != "" && ftpRespCode == FC_RESP_CODE_ENTER_PASSIVE_MODE {
					this.sendCmdRequest([]string{FC_LIST, remoteDir})
					go this.recvCmdResponse()
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
		}
	} else {
		this.cmdUsage(this.Name)
	}
}

func (this *GoFtpClientCmd) pasv() (pasvHost string, pasvPort int, ftpRespCode int, err error) {
	if this.Connected {
		this.sendCmdRequest([]string{FC_PASV})
		var recvData = this.recvCmdResponse()
		var startIndex = strings.Index(recvData, "(")
		var endIndex = strings.LastIndex(recvData, ")")
		if startIndex == -1 || endIndex == -1 {
			err = errors.New("ftp: PASV command failed.")
		} else {
			var pasvDataStr = recvData[startIndex+1 : endIndex]
			var pasvDataParts = strings.Split(pasvDataStr, ",")
			pasvHost = strings.Join(pasvDataParts[:4], ".")
			var p1, p2 int
			p1, err = strconv.Atoi(pasvDataParts[4])
			p2, err = strconv.Atoi(pasvDataParts[5])
			pasvPort = p1*256 + p2

			ftpRespCode, err = this.parseCmdResponse(recvData)
		}
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
		pasvRespData = make([]byte, 0)
		for {
			line, err := bReader.ReadBytes('\n')
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
