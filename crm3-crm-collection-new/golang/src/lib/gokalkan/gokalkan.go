// +build linux

package gokalkan

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/lestrrat-go/libxml2/xpath"

	"github.com/lestrrat-go/libxml2"
)

//#cgo LDFLAGS:-ldl
// #include <string.h>
// #include <stdio.h>
// #include <dlfcn.h>
// #include "KalkanCrypt.h"
//
// typedef int (*KC_GetFunctionList1)(stKCFunctionsType **KCfunc);
// int skipfl = 0;
// stKCFunctionsType* libInit(){
// KC_GetFunctionList1 lib_funcList = NULL;
// stKCFunctionsType *kc_funcs;
// void    *handle;
//      char* msgstr;
//      int msglen;
//          handle = dlopen("libkalkancryptwr-64.so",  RTLD_LAZY|RTLD_GLOBAL);
//          if (!handle) {
//             msgstr=dlerror();
//             printf("\ndlerror:%s\n",msgstr);
//             return NULL;
//          }
//          lib_funcList = (KC_GetFunctionList1)dlsym(handle, "KC_GetFunctionList");
//          lib_funcList(&kc_funcs);
//          unsigned int rv = kc_funcs->KC_Init();
//          if (rv) {
//             printf("\nKC_Init error\n");
//             return NULL;
//          }
//      return kc_funcs;
//
//}
//
//
// int libLoadKey( stKCFunctionsType* kc_funcs,
//         char* container,
//         char* password,
//         char* pMsg,
//         int* pMsgLen
//     ){
//     unsigned long storage = KCST_PKCS12;
//     int containerLen = strlen(container);
//     int passwordLen = strlen(password);
//     char*  alias = (char*)"sha256";
//     //printf("\ncontainer:%s %d\n",container,containerLen);
//     //printf("\npassw:%s %d\n",password,passwordLen);
//     //printf("\nmsglen:%d\n",(*pMsgLen));
//     unsigned int rv = kc_funcs->KC_LoadKeyStore(
//         storage,
//         (char*)password,
//         passwordLen,
//         (char*)container,
//         containerLen,
//         (char*)alias
//     );
//     rv = kc_funcs->KC_GetLastErrorString(pMsg, pMsgLen);
//     if(rv>0){
//        printf("\nKC_LoadKeyStore error:%s\n", pMsg);
//        return 1;
//     }
//      return 0;
//}
// int libSignXml(stKCFunctionsType* kc_funcs,
//         char* inXMLData,
//         char* signNodeId,
//         char* parentNameSpace,
//         char* parentSignNode,
//         char* pOut,
//         int* pOutLen,
//         char* pMsg,
//         int* pMsgLen
//     ){
//     int inXMLDataLength = strlen(inXMLData);
//     char*  alias = (char*)"sha256";
//     unsigned int rv = kc_funcs->SignXML(
//         (char *)alias,
//         0,
//         inXMLData,
//         inXMLDataLength,
//         (unsigned char*)pOut,
//         pOutLen,
//         (char *)signNodeId,
//         (char *)parentSignNode,
//         (char *)parentNameSpace
//     );
//     rv = kc_funcs->KC_GetLastErrorString(pMsg, pMsgLen);
//     if(rv>0){
//         printf("\nSignXML error:%s\n", pMsg);
//         return 1;
//     }
//     kc_funcs->KC_XMLFinalize();
//     return 0;
// }
// int libTSASetUrl(stKCFunctionsType* kc_funcs,char* tsaurl){
//      kc_funcs->KC_TSASetUrl((char*)tsaurl);
//}
// int libVerifyXml(stKCFunctionsType* kc_funcs,
//         char* inXMLSign,
//         char* pMsg,
//         int* pMsgLen
//     ){
//     //printf("\nlibVerifyXml start:%s\n--------------------\n", inXMLSign);
//     int inXMLSignLength = strlen(inXMLSign);
//     int outVerifyInfoLen = 8192;
//     char outVerifyInfo[outVerifyInfoLen ];
//     char*  alias = (char*)"";
//     //printf("\nlibVerifyXml alias:%s  len:%dend\n",alias,inXMLSignLength);
//     unsigned int rv= kc_funcs->VerifyXML((char *)alias,
//      0,
//      inXMLSign,
//      inXMLSignLength,
//      &outVerifyInfo[0],
//       &outVerifyInfoLen
//     );
//     //printf("\nlibVerifyXml verify end\n");
//     rv = kc_funcs->KC_GetLastErrorString(pMsg, pMsgLen);
//     if(rv>0){
//         printf("\nVerifyXml error:%s\n", pMsg);
//         return 1;
//     }
//     //printf("\nVerifyXml ok:%s\n", outVerifyInfo);
//     return 0;
// }
// int libGetCertFromXml(stKCFunctionsType* kc_funcs,
//         char* inXMLData,
//         char* pOut,
//         int* pOutLen,
//         char* pMsg,
//         int* pMsgLen
//     ){
//     int inXMLDataLength = strlen(inXMLData);
// int inSignId = 1;
// unsigned int rv = kc_funcs->KC_getCertFromXML((const char*)inXMLData,
// inXMLDataLength,
//   inSignId,
//   pOut,
//   pOutLen);
//     rv = kc_funcs->KC_GetLastErrorString(pMsg, pMsgLen);
//     if(rv>0){
//         printf("\nSignXML error:%s\n", pMsg);
//         return 1;
//     }
//     kc_funcs->KC_XMLFinalize();
//     return 0;
// }
// int libX509LoadCertificateFromFile( stKCFunctionsType* kc_funcs,
//         char* container,
//         int certType,
//         char* pMsg,
//         int* pMsgLen
//     ){
//     unsigned int rv = kc_funcs->X509LoadCertificateFromFile(
//         (char*)container,
//         certType
//     );
//     rv = kc_funcs->KC_GetLastErrorString(pMsg, pMsgLen);
//     if(rv>0){
//        printf("\nKC_LoadKeyStore error:%s\n", pMsg);
//        return 1;
//     }
//      return 0;
//}
// int libX509CertificateGetInfo( stKCFunctionsType* kc_funcs,
//         char* inCert,
//         int propId,
//         char* pOut,
//         int* pOutLen
//     ){
//     int inCertLength = strlen(inCert);
//     unsigned int rv = kc_funcs->X509CertificateGetInfo(
//         inCert,
//         inCertLength,
//         propId,
//         (unsigned char*)pOut,
//         pOutLen
//     );
//     if(rv>0){
//        printf("\nKC_X509CertificateGetInfo error\n");
//        return 1;
//     }
//      return 0;
//}
import "C"

var once sync.Once
var kc_funcs *C.stKCFunctionsType

func start() {
	once.Do(func() {
		go mainLoop()
	})
}

func mainLoop() {
	kc_funcs = C.libInit()
	for f := range mainfunc {
		f()
	}
}

var mainfunc = make(chan func())

func do(f func()) {
	start()
	done := make(chan bool, 1)
	mainfunc <- func() {
		f()
		done <- true
	}
	<-done
}

var oncekey sync.Once

func LoadKey(container string, password string) (reterr error) {
	defer func() {
		if r := recover(); r != nil {
			reterr = errors.New("exception")
		}
	}()
	msgLen := C.int(65534)
	msg := make([]byte, msgLen)
	ret := C.int(2)
	do(func() {
		oncekey.Do(func() {
			ret = C.libLoadKey(kc_funcs,
				C.CString(container),
				C.CString(password),
				(*C.char)(unsafe.Pointer(&msg[0])),
				&msgLen)
		})

	})
	if ret == 2 {
		return errors.New("error:keys already loaded")
	}
	if ret == 0 {
		return nil
	}
	return errors.New(C.GoString((*C.char)(unsafe.Pointer(&msg[0]))))

}

func signXmlInternal(xmldata []byte) (retdata []byte, reterr error) {
	defer func() {
		if r := recover(); r != nil {
			retdata = nil
			reterr = errors.New("exception")
		}
	}()
	dataLen := C.int(len(xmldata) + 50000)
	data := make([]byte, dataLen)
	msgLen := C.int(65534)
	msg := make([]byte, msgLen)
	ret := C.int(0)
	do(func() {
		ret = C.libSignXml(kc_funcs,
			(*C.char)(unsafe.Pointer(&xmldata[0])),
			C.CString(""),
			C.CString(""),
			C.CString(""),
			(*C.char)(unsafe.Pointer(&data[0])),
			&dataLen,
			(*C.char)(unsafe.Pointer(&msg[0])),
			&msgLen)
	})
	if ret == 0 {
		return data[:dataLen], nil
	}
	return nil, errors.New(C.GoString((*C.char)(unsafe.Pointer(&msg[0]))))

}
func SignXml(xmldata []byte, signNodeXpath string) (retdata []byte, reterr error) {
	if signNodeXpath == "" {
		signedText, err := signXmlInternal(xmldata)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("signXmlInternal 1 error: %v", err))
		}
		signedDoc, err := libxml2.ParseString(string(signedText))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("libxml2.ParseString(string(signedText)): %v", err))
		}
		signedNode, err := signedDoc.DocumentElement()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("signedDoc.DocumentElement(): %v", err))
		}
		return []byte(signedNode.String()), nil
	}
	dom, err := libxml2.ParseString(string(xmldata))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("libxml2.ParseString: %v", err))
	}
	xpathResult, err := dom.Find(signNodeXpath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("dom.Find: %v", err))
	}
	signNode := xpathResult.NodeList().First()
	if signNode == nil {
		return nil, errors.New("gokalkan:empty signNodeXpath result")
	}
	nodeText := signNode.String()
	signedText, err := signXmlInternal([]byte(nodeText))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("signXmlInternal: %v", err))
	}
	parent, err := signNode.ParentNode()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("signNode.ParentNode(): %v", err))
	}
	err = parent.RemoveChild(signNode)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("parent.RemoveChild(signNode): %v", err))
	}
	signedDoc, err := libxml2.ParseString(string(signedText))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("libxml2.ParseString(string(signedText)): %v", err))
	}
	signedNode, err := signedDoc.DocumentElement()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("signedDoc.DocumentElement(): %v", err))
	}
	err = parent.AddChild(signedNode)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("parent.AddChild(signedNode): %v", err))
	}
	docElement, err := dom.DocumentElement()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("dom.DocumentElement(): %v", err))
	}
	return []byte(docElement.String()), nil

}
func TSASetUrl(url string) (reterr error) {
	defer func() {
		if r := recover(); r != nil {
			reterr = errors.New("exception")
		}
	}()
	ret := C.int(0)
	do(func() {
		oncekey.Do(func() {
			ret = C.libTSASetUrl(kc_funcs,
				C.CString(url))
		})

	})
	return nil
}
func verifyXmlInternal(xmldata []byte) (reterr error) {
	defer func() {
		if r := recover(); r != nil {
			reterr = errors.New("exception")
		}
	}()

	msgLen := C.int(65534)
	msg := make([]byte, msgLen)
	ret := C.int(0)
	do(func() {
		ret = C.libVerifyXml(kc_funcs,
			(*C.char)(unsafe.Pointer(&xmldata[0])),
			(*C.char)(unsafe.Pointer(&msg[0])),
			&msgLen)
	})
	if ret == 0 {
		return nil
	}
	return errors.New(C.GoString((*C.char)(unsafe.Pointer(&msg[0]))))
}

func extraxtSignXml(xmldata []byte) (xmlsign []byte, reterr error) {
	dom, err := libxml2.ParseString(string(xmldata))
	if reterr != nil {
		return nil, errors.New(fmt.Sprintf("libxml2.ParseString: %v", err))
	}
	ctx, _ := xpath.NewContext(dom)
	_ = ctx.RegisterNS("ds", "http://www.w3.org/2000/09/xmldsig#")
	xpathResult, err := ctx.Find("//ds:Signature")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("dom.Find: %v", err))
	}
	signNode := xpathResult.NodeList().First()
	if signNode == nil {
		return nil, errors.New("gokalkan:empty signNodeXpath result")
	}
	parent, err := signNode.ParentNode()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("signNode.ParentNode: %v", err))
	}
	nodeText := parent.String()
	return []byte(nodeText), nil
}
func VerifyXml(xmldata []byte) (reterr error) {
	signxml, err := extraxtSignXml(xmldata)
	if err != nil {
		return err
	}
	return verifyXmlInternal(signxml)
}

func getCertFromXmlInternal(xmldata []byte) (retdata []byte, reterr error) {
	defer func() {
		if r := recover(); r != nil {
			retdata = nil
			reterr = errors.New("exception")
		}
	}()
	dataLen := C.int(32768)
	data := make([]byte, dataLen)
	msgLen := C.int(65534)
	msg := make([]byte, msgLen)
	ret := C.int(0)
	do(func() {
		ret = C.libGetCertFromXml(kc_funcs,
			(*C.char)(unsafe.Pointer(&xmldata[0])),
			(*C.char)(unsafe.Pointer(&data[0])),
			&dataLen,
			(*C.char)(unsafe.Pointer(&msg[0])),
			&msgLen)
	})
	if ret == 0 {
		return data[:dataLen], nil
	}
	return nil, errors.New(C.GoString((*C.char)(unsafe.Pointer(&msg[0]))))
}

func GetCertFromXml(xmldata []byte) (retdata []byte, reterr error) {
	signxml, err := extraxtSignXml(xmldata)
	if err != nil {
		return nil, err
	}
	return getCertFromXmlInternal(signxml)
}

func LoadCertificateFromFile(filename string, certType int) (reterr error) {
	defer func() {
		if r := recover(); r != nil {
			reterr = errors.New("exception")
		}
	}()
	msgLen := C.int(65534)
	msg := make([]byte, msgLen)
	ret := C.int(0)
	do(func() {
		ret = C.libX509LoadCertificateFromFile(kc_funcs,
			(*C.char)(unsafe.Pointer(&([]byte(filename))[0])),
			C.int(certType),
			(*C.char)(unsafe.Pointer(&msg[0])),
			&msgLen)
	})
	if ret == 0 {
		return nil
	}
	return errors.New(C.GoString((*C.char)(unsafe.Pointer(&msg[0]))))
}

func CertificateGetInfo(inCert []byte, propId int) (retdata []byte, reterr error) {
	defer func() {
		if r := recover(); r != nil {
			retdata = nil
			reterr = errors.New("exception")
		}
	}()
	dataLen := C.int(32768)
	data := make([]byte, dataLen)
	ret := C.int(0)
	do(func() {

		ret = C.libX509CertificateGetInfo(kc_funcs,
			(*C.char)(unsafe.Pointer(&inCert[0])),
			C.int(propId),
			(*C.char)(unsafe.Pointer(&data[0])),
			&dataLen)
	})
	if ret == 0 {
		return data[:dataLen], nil
	}
	return nil, errors.New(fmt.Sprintf("Error:%d", ret))
}
