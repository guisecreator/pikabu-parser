package main

import (
	"fmt"
	"time"

	"github.com/guisecreator/pikabu-parser/internal"
)

func main() {
	Url := "https://pikabu.ru"

	newParser := parser.New(Url)

	fmt.Printf("Parsing has begun...")

	posts := newParser.GetPosts()
	if len(posts) == 0 {
		fmt.Printf("%s:"+" No posts\n", time.Now().Format(time.UnixDate))
		return
	}

	for _, post := range posts {
		fmt.Println(post)
	}

	fmt.Printf("%s:"+" Parsing completed\n", time.Now().Format(time.UnixDate))

}
