package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var UserDir = flag.String("d", "./", "请输入目录名称") //默认是当前目录
	flag.Parse()
	filesList, err := foreachDir(*UserDir)
	if err == nil {
		sum := 0
		for _, f := range filesList {
			if strings.HasSuffix(f, ".go") {
				n, e := ReadFilesNums(f)
				if e == nil {
					fmt.Printf("文件:%s共有%d行\n", f, n)
					sum += n
				} else {
					fmt.Println(e.Error())
				}
			}

		}
		fmt.Printf("项目所有代码%d", sum)

	} else {
		fmt.Println(err.Error())
	}
}

//循环目录文件
func foreachDir(dirpath string) (fileslist []string, err error) {
	fileslist = make([]string, 5)
	num := 0
	filepath.Walk(dirpath,
		func(path string, f os.FileInfo, err error) error {

			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			fileslist[num] = path
			num++
			return nil
		})
	return fileslist, nil
}

//读取文件遍历出行数
func ReadFilesNums(file string) (num int, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	bufioRead := bufio.NewReader(f)
	num = 0
	note := false
Me:
	for {

		input, ferr := bufioRead.ReadString('\n')
		if ferr == io.EOF {
			return
		}
		if x, _ := regexp.MatchString(`^//`, input); x { //去除//注释统计
			goto Me
		} else if a, _ := regexp.MatchString(`^/*.*$\*/`, input); a { //去除/**/注释
			goto Me
		} else if c, _ := regexp.MatchString(`^/\*`, input); c { //去除多行/**/注释
			note = true
			goto Me
		} else if note == true {
			if m, _ := regexp.MatchString(`.*$*/`, input); !m { //去除多行/**/注释
				note = false
				goto Me
			}
		} else if strings.TrimSpace(input) == "" {
			//fmt.Println(input)
			goto Me
		} else {
			num++
		}
	}
	return num, err
}
