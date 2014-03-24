package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//定义gocount命令的参数
// -d 参数 指定项目的根目录，默认为命令启动时的目录
// -c 参数 指定是否需要计算注释的行数，默认为计算
var (
	dirFlag = flag.String("d", ".", "项目根目录")
	cmtFlag = flag.Bool("c", true, "包含注释")
)

//定义go语言的注释标记和文件后缀名
const (
	LINE_COMMENT_START_FLAG  string = `//`
	BLOCK_COMMENT_START_FLAG string = `/*`
	BLOCK_COMMENT_END_FLAG   string = `*/`

	GO_FILE_SUFFIX string = ".go"
)

//为每个文件定义一个GoFileLineCount结构体
type GoFileLineCount struct {
	CodeCount         int //代码行数
	LineCommentCount  int //行注释行数
	BlockCommentCount int //块注释行数
}

//获取该文件的内容总行数
func (this GoFileLineCount) totalLines() int {
	return this.CodeCount + this.LineCommentCount + this.BlockCommentCount
}

func countProject(projectDir string, withComment bool) {
	fileInfo, err := os.Stat(projectDir)
	if err == nil {
		if fileInfo.IsDir() {
			var totalCodeCount int = 0
			var totalLineCommentCount int = 0
			var totalBlockCommentCount int = 0

			filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) (retError error) {
				fileInfo, fileInfoErr := os.Stat(path)
				if fileInfoErr == nil {
					if !fileInfo.IsDir() && strings.HasSuffix(strings.ToLower(path), GO_FILE_SUFFIX) {
						fileLineCount, countErr := countFile(path)
						if countErr != nil {
							fmt.Println(path, ": 计算错误")
						} else {
							totalCodeCount += fileLineCount.CodeCount
							totalLineCommentCount += fileLineCount.LineCommentCount
							totalBlockCommentCount += fileLineCount.BlockCommentCount
						}
					}
				} else {
					retError = fileInfoErr
					fmt.Println(path, ": 无法访问")
				}
				return
			})

			fmt.Println()
			if withComment {
				fmt.Println("行注释:", totalLineCommentCount)
				fmt.Println("块注释:", totalBlockCommentCount)
			}
			fmt.Println("代码:", totalCodeCount)
			if withComment {
				totalFileLines := totalCodeCount + totalLineCommentCount + totalBlockCommentCount
				totalCommentLines := totalBlockCommentCount + totalLineCommentCount
				fmt.Println("总计:", totalFileLines)
				var commentPercent = float64(totalCommentLines) / float64(totalFileLines) * 100
				var codePercent = float64(totalCodeCount) / float64(totalFileLines) * 100
				fmt.Printf("注释比例:%.2f%% 代码比例:%.2f%%", commentPercent, codePercent)
				fmt.Println()
			}
		} else {
			fmt.Println("无法访问目录:", projectDir)
		}
	} else {
		fmt.Println(err)
	}
}

//获取每个文件的内容行数
func countFile(path string) (fileLineCount GoFileLineCount, err error) {
	file, fileError := os.Open(path)
	if fileError == nil {
		fileLineCount = GoFileLineCount{}
		bReader := bufio.NewReader(file)
		var isBlockComment bool = false
		for {
			lineData, readErr := bReader.ReadString('\n')
			lineData = strings.TrimSpace(lineData)
			//分别计算文件中代码，块注释和行注释的总行数
			if strings.HasPrefix(lineData, LINE_COMMENT_START_FLAG) {
				fileLineCount.LineCommentCount++
			} else if strings.HasPrefix(lineData, BLOCK_COMMENT_START_FLAG) {
				isBlockComment = true
				fileLineCount.BlockCommentCount++
			} else if strings.HasPrefix(lineData, BLOCK_COMMENT_END_FLAG) {
				isBlockComment = false
				fileLineCount.BlockCommentCount++
			} else {
				if isBlockComment {
					fileLineCount.BlockCommentCount++
				} else {
					fileLineCount.CodeCount++
				}
			}
			if readErr == io.EOF {
				break
			}
		}
		file.Close()
		fmt.Println(path, fileLineCount.totalLines())
	} else {
		err = fileError
	}

	return
}

func main() {
	flag.Parse()

	var dir = *dirFlag
	var withComment = *cmtFlag
	countProject(dir, withComment)
}
