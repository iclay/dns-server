package custom

import (
	"dns/api"
	"fmt"
	"os"
)

//用户自定义函数，当解析返回结果时会自动根据回包类型调用此类函数
func PTRHookAction(domin []string) error {
	fmt.Fprint(os.Stdout, "hook fun PTRHookAction: ")
	return api.PutWgvpnResource(domin)
}

func AHookAction(domin []string) error {
	fmt.Fprint(os.Stdout, "hook fun AHookAction: ")
	return api.PutWgvpnResource(domin)
}

func AAAAHookAction(domin []string) error {
	fmt.Fprint(os.Stdout, "hook fun AAAAHookAction: ")
	return api.PutWgvpnResource(domin)
}

func HookAction(domin []string) error {
	fmt.Fprint(os.Stdout, "hook fun HookAction: ")
	return api.PutWgvpnResource(domin)
}
