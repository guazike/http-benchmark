// regexp.go
package stringUtils

import (
	"regexp"
)

//删除代码中的//和/**/注释
func ReplaceComment(noRemarkCont string) string {
	lineRegPatten := `\/\/[^\n]*`
	blockRegPatten := `\/\*.*?\*\/`

	lineReg, _ := regexp.Compile(lineRegPatten)
	blockReg, _ := regexp.Compile(blockRegPatten)

	noRemarkCont = lineReg.ReplaceAllString(noRemarkCont, "")
	noRemarkCont = blockReg.ReplaceAllString(noRemarkCont, "")
	return noRemarkCont
}
