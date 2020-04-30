package libs

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")

func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func SizeFormat(size float64) string {
	units := []string{"Byte", "KB", "MB", "GB", "TB"}
	n := 0
	for size > 1024 {
		size /= 1024
		n += 1
	}

	return fmt.Sprintf("%.2f %s", size, units[n])
}

func IsEmail(b []byte) bool {
	return emailPattern.Match(b)
}



func GetFilename(path string) string {
	return filepath.Base(path)
}

func GetFilesize(path string) int64 {
	fileinfo, err := os.Stat(path)
	if err == nil {
		return fileinfo.Size()
	}
	return 0
}

func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		// glog.Info(err)
		return false
	}
	return true
}

func GetTarOrZipPassword(cmd string) (pwd string) {
	cmdCleanStr := strings.Replace(cmd, "\r", "", -1)
	cmdSplitArr := strings.Split(cmdCleanStr, "\n")
	if len(cmdSplitArr) > 1 {
		for _, cmdSplit := range cmdSplitArr {
			if len(cmdSplit) == 0 {
				continue
			}
			if strings.Contains(cmdSplit, "zip") || strings.Contains(cmdSplit, "ZIP") || strings.Contains(cmdSplit, "openssl") || strings.Contains(cmdSplit, "OPENSSL") {
				cmds, _ := ParseCommandLine(cmdSplit)
				for idx, argx := range cmds {
					if argx == "-P" || argx == "-k" {
						pwd = cmds[idx+1]
						break
					}
				}
			}
		}
	} else {
		subCmdArr := strings.Split(cmd, "&&")
		if len(subCmdArr) == 0 {
			return
		}
		for _, cmd := range subCmdArr {
			if strings.Contains(cmd, "zip") || strings.Contains(cmd, "ZIP") || strings.Contains(cmd, "openssl") || strings.Contains(cmd, "OPENSSL") {
				cmds, _ := ParseCommandLine(cmd)
				for idx, argx := range cmds {
					if argx == "-P" || argx == "-k" {
						pwd = cmds[idx+1]
						break
					}
				}
			}
		}
	}
	return
}

func ParseCommandLine(command string) ([]string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for i := 0; i < len(command); i++ {
		c := command[i]
		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}
		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}
		if c == '\\' {
			escapeNext = true
			continue
		}
		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}
		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}
		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}
	if state == "quotes" {
		return []string{}, errors.New(fmt.Sprintf("Unclosed quote in command line: %s", command))
	}
	if current != "" {
		args = append(args, current)
	}
	return args, nil
}
