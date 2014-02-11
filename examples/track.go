package main

import (
	"fmt"
	"github.com/timehop/go-mixpanel"
	"os"
	"strings"
)

func usage() {
	fmt.Printf("usage: %v <token> <event> <distinct_id> [prop1=val1 prop2=val2 ...]\n", os.Args[0])
}

func main() {
	if len(os.Args) < 3 {
		usage()
		return
	}

	token := os.Args[1]
	event := os.Args[2]

	var distinctId string
	if len(os.Args) >= 4 {
		distinctId = os.Args[3]
	}

	props := map[string]string{}
	if len(os.Args) >= 5 {
		for i := 4; i < len(os.Args); i++ {
			parts := strings.Split(os.Args[i], "=")
			if len(parts) == 2 {
				props[parts[0]] = parts[1]
			}
		}
	}

	mp := mixpanel.NewMixpanel(token)
	err := mp.Track(distinctId, event, props)
	if err != nil {
		fmt.Println("Error occurred:", err)
	}
}
