package utils
import (
	"log"
	"strings"
	"fmt"
	"errors"
)


func AddLogError(err error,eventCode,errText string,v ...interface{}) error{

	errType := ""
	if strings.HasPrefix(eventCode,"E") {
		errType = "ERROR"
	}else if strings.HasPrefix(eventCode,"W") {
		errType = "WARNING"
	}else if strings.HasPrefix(eventCode,"D") {
		errType = "DEBUG"
	}else if strings.HasPrefix(eventCode,"I") {
		errType = "INFO"
	}else if strings.HasPrefix(eventCode,"V") {
		errType = "VERBOSE"
	}

	if err == nil{
		str := fmt.Sprintf("~~" + errType + "~~ " + eventCode+" ::: "+errText,v...)
		log.Println(str)
		return nil
	}else {
		str := fmt.Sprintf("~~" + errType + "~~ " + eventCode + " ::: " + errText + " ::: " + err.Error(), v...)
		err = errors.New(str)
		log.Println(err.Error())
		return err
	}

}
