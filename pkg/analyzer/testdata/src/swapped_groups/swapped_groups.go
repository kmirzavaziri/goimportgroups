package main

import ( // want `File is not goimportgroups-ed`
	"fmt"
	"os"

	"regexp"

	"strings"

	"time"
)

func Nothing() {
	fmt.Println(strings.HasPrefix("a", "b"))
	fmt.Println(os.Getenv("test"))
	fmt.Println(time.Now().String())
	fmt.Println(regexp.Regexp{})
}
