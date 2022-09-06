package utils

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	htmltemplate "text/template"
	texttemplate "text/template"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	lua "github.com/Shopify/go-lua"
)

var renderFloatPrecisionMultipliers = [10]float64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
}

var renderFloatPrecisionRounders = [10]float64{
	0.5,
	0.05,
	0.005,
	0.0005,
	0.00005,
	0.000005,
	0.0000005,
	0.00000005,
	0.000000005,
	0.0000000005,
}

func RenderFloat(format string, n float64) string {
	// Special cases:
	// NaN = "NaN"
	// +Inf = "+Infinity"
	// -Inf = "-Infinity"
	if math.IsNaN(n) {
		return "NaN"
	}
	if n > math.MaxFloat64 {
		return "Infinity"
	}
	if n < -math.MaxFloat64 {
		return "-Infinity"
	}

	// default format
	precision := 2
	decimalStr := "."
	thousandStr := ","
	positiveStr := ""
	negativeStr := "-"

	if len(format) > 0 {
		// If there is an explicit format directive,
		// then default values are these:
		precision = 9
		thousandStr = ""

		// collect indices of meaningful formatting directives
		formatDirectiveChars := []rune(format)
		formatDirectiveIndices := make([]int, 0)
		for i, char := range formatDirectiveChars {
			if char != '#' && char != '0' {
				formatDirectiveIndices = append(formatDirectiveIndices, i)
			}
		}

		if len(formatDirectiveIndices) > 0 {
			// Directive at index 0:
			// Must be a '+'
			// Raise an error if not the case
			// index: 0123456789
			// +0.000,000
			// +000,000.0
			// +0000.00
			// +0000
			if formatDirectiveIndices[0] == 0 {
				if formatDirectiveChars[formatDirectiveIndices[0]] != '+' {
					panic("RenderFloat(): invalid positive sign directive")
				}
				positiveStr = "+"
				formatDirectiveIndices = formatDirectiveIndices[1:]
			}

			// Two directives:
			// First is thousands separator
			// Raise an error if not followed by 3-digit
			// 0123456789
			// 0.000,000
			// 000,000.00
			if len(formatDirectiveIndices) == 2 {
				if (formatDirectiveIndices[1] - formatDirectiveIndices[0]) != 4 {
					panic("RenderFloat(): thousands separator directive must be followed by 3 digit-specifiers")
				}
				thousandStr = string(formatDirectiveChars[formatDirectiveIndices[0]])
				formatDirectiveIndices = formatDirectiveIndices[1:]
			}

			// One directive:
			// Directive is decimal separator
			// The number of digit-specifier following the separator indicates wanted precision
			// 0123456789
			// 0.00
			// 000,0000
			if len(formatDirectiveIndices) == 1 {
				decimalStr = string(formatDirectiveChars[formatDirectiveIndices[0]])
				precision = len(formatDirectiveChars) - formatDirectiveIndices[0] - 1
			}
		}
	}

	// generate sign part
	var signStr string
	if n >= 0.000000001 {
		signStr = positiveStr
	} else if n <= -0.000000001 {
		signStr = negativeStr
		n = -n
	} else {
		signStr = ""
		n = 0.0
	}

	// split number into integer and fractional parts
	intf, fracf := math.Modf(n + renderFloatPrecisionRounders[precision])

	// generate integer part string
	intStr := strconv.Itoa(int(intf))

	// add thousand separator if required
	if len(thousandStr) > 0 {
		for i := len(intStr); i > 3; {
			i -= 3
			intStr = intStr[:i] + thousandStr + intStr[i:]
		}
	}

	// no fractional part, we can leave now
	if precision == 0 {
		return signStr + intStr
	}

	// generate fractional part
	fracStr := strconv.Itoa(int(fracf * renderFloatPrecisionMultipliers[precision]))
	// may need padding
	if len(fracStr) < precision {
		fracStr = "000000000000000"[:precision-len(fracStr)] + fracStr
	}

	return signStr + intStr + decimalStr + fracStr
}

func TemplateGt(str1, str2 string) bool {

	f1, _ := strconv.ParseFloat(str1, 64)
	f2, _ := strconv.ParseFloat(str2, 64)
	return f1 > f2
}

func TemplateContains(str1, str2 string) bool {
	return strings.Contains(str1, str2)
}
func (context *TemplateContext) TemplateParseTemplate(s interface{}, arr interface{}) string {
	if s == nil {
		return ""
	}
	res, _ := ParseTemplate(context.Lua, s.(string), arr, context.UserId)
	return res
}

func (context *TemplateContext) TemplateLua(strFunc string, param1 interface{}) string {

	//script, _ := context.Lua.ToString(1)
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param1))
	context.Lua.SetGlobal("param1")

	err := lua.DoString(context.Lua, context.Funcs[strFunc].(string))
	if err != nil {
		return ""
	}
	value, _ := context.Lua.ToString(4)
	return value
}

func (context *TemplateContext) TemplateLua2(strFunc string, param1, param2 interface{}) string {

	//script, _ := context.Lua.ToString(1)
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param1))
	context.Lua.SetGlobal("param1")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param2))
	context.Lua.SetGlobal("param2")

	err := lua.DoString(context.Lua, context.Funcs[strFunc].(string))
	if err != nil {
		return ""
	}
	value, _ := context.Lua.ToString(4)
	return value
}

func (context *TemplateContext) TemplateLua3(strFunc string, param1, param2, param3 interface{}) string {

	//script, _ := context.Lua.ToString(1)
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param1))
	context.Lua.SetGlobal("param1")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param2))
	context.Lua.SetGlobal("param2")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param3))
	context.Lua.SetGlobal("param3")
	err := lua.DoString(context.Lua, context.Funcs[strFunc].(string))
	if err != nil {
		return ""
	}
	value, _ := context.Lua.ToString(4)
	return value
}

func (context *TemplateContext) TemplateLua4(strFunc string, param1, param2, param3, param4 interface{}) string {

	//script, _ := context.Lua.ToString(1)
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param1))
	context.Lua.SetGlobal("param1")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param2))
	context.Lua.SetGlobal("param2")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param3))
	context.Lua.SetGlobal("param3")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param4))
	context.Lua.SetGlobal("param4")
	err := lua.DoString(context.Lua, context.Funcs[strFunc].(string))
	if err != nil {
		return ""
	}
	value, _ := context.Lua.ToString(4)
	return value
}

func (context *TemplateContext) TemplateLua5(strFunc string, param1, param2, param3, param4, param5 interface{}) string {

	//script, _ := context.Lua.ToString(1)
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param1))
	context.Lua.SetGlobal("param1")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param2))
	context.Lua.SetGlobal("param2")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param3))
	context.Lua.SetGlobal("param3")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param4))
	context.Lua.SetGlobal("param4")
	context.Lua.PushLightUserData(fmt.Sprintf("%v", param5))
	context.Lua.SetGlobal("param5")
	err := lua.DoString(context.Lua, context.Funcs[strFunc].(string))
	if err != nil {
		return ""
	}
	value, _ := context.Lua.ToString(4)
	return value
}

func TemplateNEq(str1, str2 interface{}) bool {
	return fmt.Sprintf("%v", str1) != fmt.Sprintf("%v", str2)
}

func TemplateIfNil(str1, str2 interface{}) interface{} {
	if str1 == nil {
		return str2
	}
	return str1
}

func TemplateIsNil(str1 interface{}) bool {
	return str1 == nil
}

func TemplateEq(str1, str2 interface{}) bool {
	return fmt.Sprintf("%v", str1) == fmt.Sprintf("%v", str2)
}
func TemplateString(str interface{}) string {
	return "1"
}
func TemplateInitMapValue(str map[string]interface{}, code string, value interface{}) string {
	if str[code] == nil {
		fmt.Println("TemplateInitX setting value ", value)
		str[code] = value
	}
	return ""
}
func TemplateSetMapValue(str map[string]interface{}, code string, value interface{}) string {
	str[code] = value
	fmt.Println("TemplateSetX( setting value ", value)
	return ""
}

type TemplateContext struct {
	Funcs  orm.Params
	Lua    *lua.State
	UserId int64
}

func ParseTemplate(l *lua.State, s string, a interface{}, userId int64) (string, error) {

	o := orm.NewOrm()
	o.Using("default")

	context := TemplateContext{}
	context.Lua = l
	context.UserId = userId
	_, err := o.Raw(DbBindReplace(`select code as "code",script  as "script" from template_funcs`)).RowsToMap(&context.Funcs, "code", "script")
	if err != nil {
		log.Println("ParseTemplate Warning! ", err)
		//return "", nil
		//return "", err

	} 
	buf := new(bytes.Buffer)
	//fmt.Println("ParseTemplate", userId)

	funcMap := texttemplate.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"inc": func(i, i2 int) int {
			return i + i2
		},
		"parseTemplate": context.TemplateParseTemplate,
		"lua":           context.TemplateLua,
		"lua2":          context.TemplateLua2,
		"lua3":          context.TemplateLua3,
		"lua4":          context.TemplateLua4,
		"lua5":          context.TemplateLua5,
		"title":         strings.Title,
		"eq":            TemplateEq,
		"isnil":         TemplateIsNil,
		"ifnil":         TemplateIfNil,
		"neq":           TemplateNEq,
		"gt":            TemplateGt,
		"contains":      TemplateContains,
		"string":        TemplateString,
		"setMapValue":   TemplateSetMapValue,
		"initMapValue":  TemplateInitMapValue,
		"ruNum2Word":    RuNum2Word,
		"getParamValue": GetParamValue,
		"getUserParamValue": func(param string) string {
			fmt.Println("getUserParamValue,", userId, ","+param, "!")
			return GetUserParamValue(o, userId, param)
		},
		"formatNumber": func(value interface{}) string {
			if value == nil {
				return ""
			}
			v, _ := strconv.ParseFloat(value.(string), 64)
			s = fmt.Sprintf(RenderFloat("#\u202F###.##", v))
			return s
		},
	}

	t, err := texttemplate.New("test").Funcs(funcMap).Parse(s)
	if err != nil {
		fmt.Println("ParseTemplate error ", err)
		return "", err
	}
	err = t.Execute(buf, a)
	if err != nil {
		fmt.Println("ParseTemplate error ", err)
		return "", err
	}
	return string(buf.Bytes()), err
}

func ParseHTMLTemplate(l *lua.State, s string, a interface{}, userId int64) (string, error) {

	o := orm.NewOrm()
	o.Using("default")

	context := TemplateContext{}
	context.Lua = l
	context.UserId = userId
	_, err := o.Raw(DbBindReplace(`select code as "code",script  as "script" from template_funcs`)).RowsToMap(&context.Funcs, "code", "script")
	if err != nil {
		log.Println("ParseTemplate Warning! ", err)
		//return "", nil
		//return "", err

	} else {
		log.Println("ParseTemplate", context.Funcs)
	}
	buf := new(bytes.Buffer)
	fmt.Println("ParseTemplate", userId)

	funcMap := htmltemplate.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"inc": func(i, i2 int) int {
			return i + i2
		},
		"parseTemplate": context.TemplateParseTemplate,
		"lua":           context.TemplateLua,
		"lua2":          context.TemplateLua2,
		"lua3":          context.TemplateLua3,
		"lua4":          context.TemplateLua4,
		"lua5":          context.TemplateLua5,
		"title":         strings.Title,
		"eq":            TemplateEq,
		"isnil":         TemplateIsNil,
		"ifnil":         TemplateIfNil,
		"neq":           TemplateNEq,
		"gt":            TemplateGt,
		"contains":      TemplateContains,
		"string":        TemplateString,
		"setMapValue":   TemplateSetMapValue,
		"initMapValue":  TemplateInitMapValue,
		"ruNum2Word":    RuNum2Word,
		"getParamValue": GetParamValue,
		"getUserParamValue": func(param string) string {
			fmt.Println("getUserParamValue,", userId, ","+param, "!")
			return GetUserParamValue(o, userId, param)
		},
		"formatNumber": func(value interface{}) string {
			if value == nil {
				return ""
			}
			v, _ := strconv.ParseFloat(value.(string), 64)
			s = fmt.Sprintf(RenderFloat("#\u202F###.##", v))
			return s
		},
	}

	t, err := htmltemplate.New("test").Funcs(funcMap).Parse(s)
	if err != nil {
		fmt.Println("ParseTemplate error ", err)
		return "", err
	}
	err = t.Execute(buf, a)
	if err != nil {
		fmt.Println("ParseTemplate error ", err)
		return "", err
	}
	return string(buf.Bytes()), err
}
