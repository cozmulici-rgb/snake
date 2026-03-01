//go:build !windows

package graphic

import "fmt"

func Run() {
	fmt.Println("Graphic mode is supported on Windows for this project.")
}
