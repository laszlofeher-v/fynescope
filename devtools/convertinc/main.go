package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"
)

/*
read file to array of bytes
for

	find next #define line
	if not vound break
	get name
	get comments above
	if found format comment to single line, remove all //
	else single line comment=""
	write name":" "name single line comment"
*/
const (
	comment = "//"
	define  = "#define"
)

func main() {
	data, err := ioutil.ReadFile("/opt/picoscope/include/libps2000a/PicoStatus.h")
	if err != nil {
		log.Println(err)
		return
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if bytes.Contains([]byte(line), []byte(define)) {
			// fmt.Println(line)
			ids := bytes.Fields([]byte(line))
			// fmt.Println(len(id))
			// fmt.Println(string(id[0]))
			// fmt.Println(string(id[1]))
			if i > 0 && bytes.Contains([]byte(lines[i-1]), []byte(comment)) {
				j := i - 1
				for j >= 0 && bytes.Contains([]byte(lines[j]), []byte(comment)) {
					j--
				}
				log.Print("C." + string(ids[1]) + ":" + `"` + string(ids[1]) + " ")
				for j < i {
					comment := bytes.TrimLeft([]byte(lines[j]), "//")
					log.Print(string(comment))
					j++
				}
				log.Println(`",`)
				// if "PICO_DEVICE_SAMPLING" == string(ids[1]) {
				// 	return
				// }
				//fmt.Println(line)
			} else {
				log.Println("C." + string(ids[1]) + ":" + `"` + string(ids[1]) + `",`)
			}
		}
	}
}
