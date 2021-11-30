package svc

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

//GConf - struct GConf
// label标签名字必须要与配置文件一致,否则无法解析,支持int和string类型
type GConf struct {
	Include         string `label:"include" parse_func:"parse_file"`
	AccountType     string `label:"account_type"`
	GrantType       string `label:"grant_type"`
	Email           string `label:"email"`
	RemoteHost      string `label:"remote_host"`
	Password        string `label:"password"`
	LogPath         string `label:"log_path"`
	LogLevel        string `label:"log_level"` // trace, debug, info, warn[ing], error, fatal, panic
	LogMaxDiskUsage int64  `label:"log_max_disk_usage" parse_func:"parse_bytes"`
	LogMaxFileNum   int    `label:"log_max_file_num"`
	RWDirPath       string `label:"rw_path"`
	ForwardIP       string `label:"forward_ip"`
	ForwardPort     int    `label:"forward_port"`
	ServerPort      int    `label:"server_port"`

	WhiteList string `label:"white_list"` //白名单目录
	// BlackList string `label:"black_list"` //黑名单目录
}

//
var (
	GCONF     GConf
	GConfItem = &GCONF
)

//ParseBool - ParseBool
func ParseBool(value string) int {
	if value == "yes" || value == "on" || value == "1" {
		return 1
	}
	return 0
}

// ParseAsBytes parse string like 2B, 1M, 1G to bytes
func ParseAsBytes(value string) int64 {
	if len(value) == 0 {
		return 0
	}

	last := value[len(value)-1]
	if last >= '0' && last <= '9' {
		i, _ := strconv.ParseInt(value, 10, 64)
		return i
	}
	first := value[:len(value)-1]
	i, _ := strconv.ParseInt(first, 10, 64)
	switch last {
	case 'b':
		return i / 8
	case 'B':
		return i
	case 'k', 'K':
		return i * 1024
	case 'M', 'm':
		return i * 1024 * 1024
	case 'G', 'g':
		return i * 1024 * 1024 * 1024
	}
	return i
}

//ParseFile - Parse File
func ParseFile(filePath string) {
	fmt.Println("parsing file", filePath)

	confile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer confile.Close()

	lineNum := 0

	scanner := bufio.NewScanner(confile)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if len(line) < 3 || line[0] == '#' || line[0] == '[' {
			continue
		}

		var key string
		var value []string
		v := strings.Fields(line)
		if len(v) < 2 {
			fmt.Println("error parseing file ", filePath, "line", lineNum, line)
			panic(err)
		}
		key = v[0]
		value = v[1:]

		object := reflect.ValueOf(GConfItem)
		myref := object.Elem()
		typeOfType := myref.Type()
		for i := 0; i < myref.NumField(); i++ {
			fieldInfo := myref.Type().Field(i)
			tag := fieldInfo.Tag
			tagName := tag.Get("label")
			parseFunc := tag.Get("parse_func")
			variableName := typeOfType.Field(i).Name

			if strings.Compare(string(key), tagName) == 0 {
				switch parseFunc {
				case "":
					if myref.FieldByName(variableName).Kind() == reflect.Int {
						*(*int)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())), _ = strconv.Atoi(value[0])
					} else if myref.FieldByName(variableName).Kind() == reflect.Int64 {
						*(*int64)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())), _ = strconv.ParseInt(value[0], 10, 64)
					} else {
						*(*string)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())) = value[0]
					}
				case "parse_file":
					ParseFile(value[0])
				case "parse_bool":
					*(*int)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())) = ParseBool(value[0])
				case "parse_string_list":
					*(*[]string)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())) = ParseStringList(value[0])
				case "parse_bytes":
					*(*int64)(unsafe.Pointer(myref.FieldByName(variableName).Addr().Pointer())) = ParseAsBytes(value[0])

				}

				break
			}
		}
	}
}

func printInterface(depth int, inter interface{}) {
	v := reflect.ValueOf(inter)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		for j := 0; j < depth; j++ {
			fmt.Printf("\t")
		}
		f := v.Field(i)
		if f.Kind() == reflect.Struct {
			fmt.Printf("[%s]\n", t.Field(i).Name)
			printInterface(depth+1, f.Interface())
		} else {
			fmt.Printf("%s %s = %v\n", t.Field(i).Name, f.Type(), f.Interface())
		}
	}
}

//PrintInterface - Print Interface
func PrintInterface(inter interface{}) {
	printInterface(0, inter)
}

//Print -
/************打印结构体内容，调试使用****************/
func Print() {
	PrintInterface(*GConfItem)
}

func ParseStringList(value string) []string {
	return strings.Split(value, ",")
}
