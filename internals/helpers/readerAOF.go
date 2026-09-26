package helpers

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func ReaderLineByLine(s *store.InMemoryStore) error {
	d, err := os.Open("access.log")
	if err != nil {
		return err
	}
	defer d.Close()
	scanner := bufio.NewReader(d)

	for {
		args, err := readCommand(scanner)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Print("running dispatch on the args: \n")
		for _, d := range args {
			fmt.Printf("%s ", d)
		}
		fmt.Print("\n")
		dispatch(args, s)
	}
	return nil
}

// func processOneLine(data []byte) error {
// 	if len(data) <= 0 {
// 		return nil
// 	}
// 	args, err := readCommand(data)
// 	if err != nil {
// 		return
// 	}

// 	ans := dispatch(args, s)
// }

// func main() {

// 	data := `127.0.0.1 - - [26/Sep/2026:10:00:00 +0000] "GET /index.html HTTP/1.1" 200 1024`
// 	err := Writter([]byte(data))
// 	if err != nil{
// 		fmt.Println("some thing not write in writter", err)
// 		return
// 	}
// 	err = ReaderLineByLine()
// 	if err != nil {
// 		fmt.Println("Error reading file:", err)
// 	}
// }
