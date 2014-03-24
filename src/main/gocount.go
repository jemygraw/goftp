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

var (
	dirFlag = flag.String("d", ".", "项目根目录")
	cmtFlag = flag.Bool("c", true, "包含注释")
)

const (
	LINE_COMMENT_START_FLAG  string = `//`
	BLOCK_COMMENT_START_FLAG string = `/*`
	BLOCK_COMMENT_END_FLAG   string = `*/`

	GO_FILE_SUFFIX string = ".go"
)

type GoFileLineCount struct {
	CodeCount         int
	LineCommentCount  int
	BlockCommentCount int
}

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
							fmt.Println(path, ": Count Error")
						} else {
							totalCodeCount += fileLineCount.CodeCount
							totalLineCommentCount += fileLineCount.LineCommentCount
							totalBlockCommentCount += fileLineCount.BlockCommentCount
						}

					}
				} else {
					retError = fileInfoErr
					fmt.Println(path, ": Access Error")
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
			fmt.Println("Can not access directory", projectDir)
		}
	} else {
		fmt.Println(err)
	}
}

func countFile(path string) (fileLineCount GoFileLineCount, err error) {
	file, fileError := os.Open(path)
	if fileError == nil {
		fileLineCount = GoFileLineCount{}
		bReader := bufio.NewReader(file)
		var isBlockComment bool = false
		for {
			lineData, readErr := bReader.ReadString('\n')
			lineData = strings.TrimSpace(lineData)

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
