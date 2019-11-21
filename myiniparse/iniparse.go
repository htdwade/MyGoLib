package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

// ini配置文件解析器

// MysqlConfig MySQL配置结构体
type MysqlConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

// RedisConfig ...
type RedisConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Password string `ini:"password"`
	Database int    `ini:"database"`
	Test     bool   `ini:"test"`
}

// Config ...
type Config struct {
	MysqlConfig `ini:"mysql"`
	RedisConfig `ini:"redis"`
}

// loadIni 解析ini配置到结构体
func loadIni(fileName string, data interface{}) (err error) {
	// 1. 参数校验
	// 1.1 传进来的data参数必须是指针类型(因为需要在函数中对其赋值修改)
	t := reflect.TypeOf(data)
	fmt.Println(t.Name(), t.Kind())
	if t.Kind() != reflect.Ptr {
		err = errors.New("data param should be a pointer")
		return
	}
	// 1.2 传进来的data参数必须是结构体类型指针(因为配置文件中各种键值对需要赋值给结构体的字段)
	if t.Elem().Kind() != reflect.Struct {
		err = errors.New("data param should be a struct pointer")
		return
	}
	// 2. 读文件得到字节类型数据
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		err = errors.New("ioutil readfile err")
		return
	}

	lineSlice := strings.Split(string(b), "\r\n")

	// 3. 一行一行的读数据
	var structName string
	for idx, line := range lineSlice {
		// 3.1 空白字符处理
		// 去掉字符串首尾的空格
		line = strings.TrimSpace(line)
		// 如果是空行就跳过
		if len(line) == 0 {
			continue
		}
		// 3.2 如果是注释就跳过
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		// 3.3 如果是[开头的就表示是节(section)
		if strings.HasPrefix(line, "[") {
			if line[len(line)-1] != ']' {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return
			}
			// 把这一行首尾的[]去掉，取到中间的内容把首尾的空格去掉拿到内容
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if len(sectionName) == 0 {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return
			}
			// 根据字符串sectionName去data里面根据反射找到对应的结构体
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if sectionName == field.Tag.Get("ini") {
					// 说明找到了对应的嵌套结构体，把字段名记下来
					structName = field.Name
					fmt.Printf("找到%s对应的嵌套结构体%s\n", sectionName, structName)
					break
				}
			}
		} else {
			// 3.4 如果不是[开头就是=分割的键值对
			// 以等号分割这一行，等号左边是key,等号右边是value
			if strings.Index(line, "=") == -1 || strings.HasPrefix(line, "=") {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return
			}
			index := strings.Index(line, "=")
			key := strings.TrimSpace(line[:index])
			value := strings.TrimSpace(line[index+1:])
			// 根据strucrName 去 data 里面把对应的嵌套结构体给取出来
			v := reflect.ValueOf(data)
			sValue := v.Elem().FieldByName(structName) // 拿到嵌套结构体的值信息
			sType := sValue.Type()                     // 拿到嵌套结构体的类型信息

			if sType.Kind() != reflect.Struct {
				err = fmt.Errorf("data中的%s字段应该是一个结构体", structName)
				return
			}
			var fieldName string
			var fileType reflect.StructField
			// 遍历嵌套结构体的每一个字段，判断tag是不是等于key
			for i := 0; i < sValue.NumField(); i++ {
				filed := sType.Field(i) // tag信息是存储在类型信息中的
				if filed.Tag.Get("ini") == key {
					// 找到对应的字段
					fieldName = filed.Name
					fileType = filed
					break
				}
			}
			// 如果key = tag，给这个字段赋值
			// 根据fieldName 去取出这个字段
			if len(fieldName) == 0 {
				// 在结构体中找不到对应的字段
				continue
			}
			fileObj := sValue.FieldByName(fieldName)
			// 对其赋值
			fmt.Println(fieldName, fileType.Type.Kind())
			switch fileType.Type.Kind() {
			case reflect.String:
				fileObj.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var valueInt int64
				valueInt, err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return
				}
				fileObj.SetInt(valueInt)
			case reflect.Bool:
				var valueBool bool
				valueBool, err = strconv.ParseBool(value)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return
				}
				fileObj.SetBool(valueBool)
			case reflect.Float32, reflect.Float64:
				var valueFloat float64
				valueFloat, err = strconv.ParseFloat(value, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return
				}
				fileObj.SetFloat(valueFloat)
			}
		}
	}
	return
}

func main() {
	var cfg Config
	err := loadIni("./conf.ini", &cfg)
	if err != nil {
		fmt.Printf("load ini failed, err:%v\n", err)
		return
	}
	fmt.Printf("%#v\n", cfg)
}
