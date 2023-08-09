package main

import ( // want `File is not goimportgroups-ed`
	"fmt"
	"os"

	"time"

	"regexp"
	"strings"
)

func Nothing() {
	fmt.Println(strings.HasPrefix("a", "b"))
	fmt.Println(os.Getenv("test"))
	fmt.Println(time.Now().String())
	fmt.Println(regexp.Regexp{})
}
