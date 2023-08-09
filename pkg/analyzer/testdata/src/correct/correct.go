package main

import (
	"fmt"
	"os"

	"time"

	"strings"

	"regexp"
)

func Nothing() {
	fmt.Println(strings.HasPrefix("a", "b"))
	fmt.Println(os.Getenv("test"))
	fmt.Println(time.Now().String())
	fmt.Println(regexp.Regexp{})
}
