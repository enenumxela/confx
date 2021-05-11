package main

import (
	"fmt"
	"log"

	"github.com/enenumxela/confx/pkg/confx"
)

func init() {
	confx.SetOverride("host", "override value")

	if err := confx.SetConfiguration("./conf.yaml"); err != nil {
		log.Fatalln(err)
	}

	confx.SetDefault("default", "default value")
}

func main() {
	fmt.Println(confx.Get("host").(string))
	fmt.Println(confx.Get("port").(int))
	fmt.Println(confx.Get("default").(string))
}
