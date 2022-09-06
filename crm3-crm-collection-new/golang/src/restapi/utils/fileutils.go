package utils

import (
	"io/ioutil"
	"log"
	"io"
	"os"
	"fmt"
	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)


func CopyFile(src, dst string) (int64, error) {
        sourceFileStat, err := os.Stat(src)
        if err != nil {
                return 0, err
        }

        if !sourceFileStat.Mode().IsRegular() {
                return 0, fmt.Errorf("%s is not a regular file", src)
        }

        source, err := os.Open(src)
        if err != nil {
                return 0, err
        }
        defer source.Close()

        destination, err := os.Create(dst)
        if err != nil {
                return 0, err
        }
        defer destination.Close()
        nBytes, err := io.Copy(destination, source)
        return nBytes, err
}

func TempFile(ext string) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	f.Close()
	os.Remove(f.Name())
	name := f.Name() + ext
	f = nil
	return name, nil
}

func UploadRawData(o orm.Ormer, dirCode, fileName string, data []byte) (string, error) {

	var uuid = ""

	dir := ""
	pth := "unix_path"
	if runtime.GOOS == "windows" {
		pth = "win_path"
	}

	restCode := ""
	restBody := ""
	filename_as_uuid := 0
	path_expr := ""

	err := o.Raw(DbBindReplace("select  path_expr, coalesce(filename_as_uuid,0),"+pth+",(select code from rest_services where id=dirs.restservice_id) restCode,(select body from rest_services where id=dirs.restservice_id) restBody from dirs where code=?"), dirCode).QueryRow(&path_expr, &filename_as_uuid, &dir, &restCode, &restBody)
	if err != nil {
		log.Println("suk3 " + err.Error())
		return "", err
	}
	uuid = dirCode + "-" + Uuid()
	fullFileName := dir + uuid
	filepath := uuid

	if filename_as_uuid == 1 {
		os.MkdirAll(dir+uuid, os.ModePerm)
		filepath = uuid + "/" + fileName
		fullFileName = dir + uuid + "/" + fileName
	} else {

	}
	if os.Getenv("CRM_DEBUG_ECM") == "1" {
		log.Println("uuid=" + uuid)
	}

	_, err = DbInsert(o, DbBindReplace("insert into files (filepath,dir_id,code,title,filename) values (?,(select id from dirs where code=?),?,?,?)"),
		filepath, dirCode, uuid, fileName, fileName)
	if err != nil {
		log.Println("suk2 " + err.Error())
		return "", err
	}

	if os.Getenv("CRM_DEBUG_ECM") == "1" {
		log.Println("fullFileName=" + fullFileName)
	}

	//err = os.MkdirAll(dir+filepath, 0666)

	if err != nil {
		log.Println("Error on Create Directory " + err.Error())
		return "", err
	}

	f, err := os.OpenFile(fullFileName, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Println("Error on OpenFile " + err.Error())
		return "", err
	}

	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Println("Error on OpenFile " + err.Error())
		return "", err
	}
	f.Close()

	return uuid, nil
}
