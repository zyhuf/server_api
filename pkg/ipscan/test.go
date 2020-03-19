package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sbinet/go-python"
)

var PyStr = python.PyString_FromString

var IPScan *python.PyObject

var funcsModule *python.PyObject
var funcName *python.PyObject

func init() {
	os.Setenv("PYTHONPATH", ".")
	fmt.Println("PYTHONPATH:", os.Getenv("PYTHONPATH"))
}

func import_and_call_func() {
	funcsModule = python.PyImport_ImportModuleNoBlock("ipscan")
	if funcsModule == nil {
		panic("[MODULE REF] Error importing module: ipscan.py")
	}
	fmt.Printf("[MODULE REF] repr(funcsModule) = %s\n", python.PyString_AS_STRING(funcsModule.Repr()))

	funcName = funcsModule.GetAttrString("scanIP")
	if funcName == nil {
		panic("[FUNCTION REF] Error importing function: scanIP")
	}
	fmt.Printf("[FUNCTION REF] repr(funcName) = %s\n", python.PyString_AS_STRING(funcName.Repr()))

	callFuncIPScan("10.2.21.1-10.2.21.10")
	callFuncIPScan("10.2.21.11-10.2.21.20")
}

func callFuncIPScan(ipSegement string) {
	methodArg := python.PyDict_New()
	err := python.PyDict_SetItem(methodArg, PyStr(ipSegement), python.PyInt_FromLong(len(ipSegement)))
	if err != nil {
		panic("[LIST SET REF] Error setting element at list[0]: methodArg")
	}

	err = python.PyDict_SetItem(methodArg, PyStr("ens33"), python.PyInt_FromLong(5))
	if err != nil {
		panic("[LIST SET REF] Error setting element at list[0]: methodArg")
	}

	funcCall := funcName.Call(python.PyTuple_New(0), methodArg)
	if funcCall == nil {
		panic("[FUNCTION CALL REF] Error calling function: scanIP")
	}
	fmt.Printf("[FUNCTION CALL REF] repr(methodCall) = %s\n", python.PyString_AS_STRING(funcCall.Repr()))
}

func import_and_use_obj() {
	// Import class from module
	objsModule := python.PyImport_ImportModule("ipscan")
	if objsModule == nil {
		panic("Error importing module: ipscan")
	}

	IPScan = objsModule.GetAttrString("IPScan")
	if IPScan == nil {
		panic("[CLASS REF] Error importing object: IPScan")
	}
	go CallIPSanMethod("10.2.21.1-10.2.21.10")
	//go CallIPSanMethod("10.2.21.11-10.2.21.20")
}

func CallIPSanMethod(ipSegement string) {
	// Instantiate obj
	ObjIPSan := python.PyInstance_New(IPScan, nil, nil)
	if ObjIPSan == nil {
		panic("[INSTANCE REF] Error instantiating object: IPScan")
	}
	// Now try with multiple args
	methodArg := python.PyList_New(1)
	// err := python.PyList_SetItem(methodArg, 0, PyStr("10.2.21.1-10.2.21.10"))
	err := python.PyList_SetItem(methodArg, 0, PyStr(ipSegement))
	if err != nil {
		panic("[LIST SET REF] Error setting element at list[0]: methodArgTwoIPv4")
	}
	methodCallArgs := ObjIPSan.CallMethodObjArgs("scan_ip_mac", methodArg)
	if methodCallArgs == nil {
		panic("[METHOD CALL REF] Error calling object method: scan_ip_mac")
	}
	fmt.Printf("[METHOD CALL REF] repr(methodCallArgs) return value = %s\n", python.PyString_AS_STRING(methodCallArgs.Repr()))
}

func main() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}
	defer python.Finalize()
	//ifimport_and_use_obj()
	import_and_call_func()

	ticker := time.NewTicker(time.Duration(2) * time.Second)
	for {
		select {
		case <-ticker.C:
		}
	}
}
