package iniconfig

import (
	"errors"
	"fmt"
	"github.com/common/constants"
	"github.com/common/model"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

var Config model.Config

func Load() error {
	if err := UnMarshalFile(constants.DefaultConfigFilePath, &Config); err != nil {
		return fmt.Errorf("unmarshal failed, :%v", err)
	}
	return nil
}

func UnMarshalFile(filename string, result interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	return UnMarshal(data, result)
}

func UnMarshal(data []byte, result interface{}) (err error) {
	lineArr := strings.Split(string(data), "\n")

	typeInfo := reflect.TypeOf(result)
	if typeInfo.Kind() != reflect.Ptr {
		return errors.New("please pass address")
	}

	typeStruct := typeInfo.Elem()
	if typeStruct.Kind() != reflect.Struct {
		return errors.New("please pass struct")
	}

	var lastFieldName string
	for index, line := range lineArr {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		//如果是注释，直接忽略
		if line[0] == ';' || line[0] == '#' {
			continue
		}

		if line[0] == '[' {
			lastFieldName, err = parseSection(line, typeStruct)
			if err != nil {
				return fmt.Errorf("%v lineno:%d", err, index+1)
			}
			continue
		}

		err = parseItem(lastFieldName, line, result)
		if err != nil {
			return fmt.Errorf("%v lineno:%d", err, index+1)
		}
	}
	return nil
}

func parseItem(lastFieldName string, line string, result interface{}) error {
	index := strings.Index(line, "=")
	if index == -1 {
		return fmt.Errorf("sytax error, line:%s", line)
	}

	key := strings.TrimSpace(line[0:index])
	val := strings.TrimSpace(line[index+1:])

	if len(key) == 0 {
		return fmt.Errorf("sytax error, line:%s", line)
	}

	resultValue := reflect.ValueOf(result)
	sectionValue := resultValue.Elem().FieldByName(lastFieldName)

	sectionType := sectionValue.Type()
	if sectionType.Kind() != reflect.Struct {
		return fmt.Errorf("field:%s must be struct", lastFieldName)
	}

	keyFieldName := ""
	for i := 0; i < sectionType.NumField(); i++ {
		field := sectionType.Field(i)
		tagVal := field.Tag.Get("ini")
		if tagVal == key {
			keyFieldName = field.Name
			break
		}
	}

	if len(keyFieldName) == 0 {
		return fmt.Errorf("KeyFieldName cannot be empty")
	}

	fieldValue := sectionValue.FieldByName(keyFieldName)
	if fieldValue == reflect.ValueOf(nil) {
		return fmt.Errorf("FieldValue cannot be nil")
	}

	switch fieldValue.Type().Kind() {
	case reflect.String:
		fieldValue.SetString(val)

	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		intVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(intVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)

	default:
		return fmt.Errorf("unsupport type:%v", fieldValue.Type().Kind())
	}

	return nil
}

func parseSection(line string, typeInfo reflect.Type) (string, error) {
	if line[0] == '[' && len(line) <= 2 {
		return "", fmt.Errorf("syntax error, invalid section:%s", line)
	}

	if line[0] == '[' && line[len(line)-1] != ']' {
		return "", fmt.Errorf("syntax error, invalid section:%s", line)
	}
	var sectionName string
	if line[0] == '[' && line[len(line)-1] == ']' {
		sectionName = strings.TrimSpace(line[1 : len(line)-1])
		if len(sectionName) == 0 {
			return "", fmt.Errorf("syntax error, invalid section:%s", line)
		}

		for i := 0; i < typeInfo.NumField(); i++ {
			field := typeInfo.Field(i)
			tagValue := field.Tag.Get("ini")
			if tagValue == sectionName {
				fieldName := field.Name
				return fieldName, nil
			}
		}
	}

	return "", fmt.Errorf("Cannot find field name by section name %s", sectionName)
}
