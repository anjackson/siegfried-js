package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"syscall/js"

	"github.com/richardlehane/siegfried"
)

func main() {
	quit := make(chan struct{}, 0)

	// Pick up the selected sig file:
	ss := js.Global().Get("sig-select")
	sigSelected := ss.Get("value").String() //.Get(ss.Get("selectedIndex").String()).Get("value").String()

	log.Print(sigSelected)

	resp, err := http.Get(sigSelected)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("NOPE")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//bodyString := string(bodyBytes)
	//log.Print(bodyString)

	s, err := siegfried.LoadReader(bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatal(err)
	}

	document := js.Global().Get("document")
	//document.Get("body").Call("insertAdjacentHTML", "beforeend", `
	//	<input type="file" id="fileInput"><br>
	//	<output id="fileOutput"></output>
	//`)

	fileInput := document.Call("getElementById", "fileInput")
	fileOutput := document.Call("getElementById", "output")

	fileInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
		fileOutput.Set("innerText", "")
		fileInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
			data := js.Global().Get("Uint8Array").New(x[0])
			dst := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(dst, data)

			filename := fileInput.Get("files").Call("item", 0).Get("name").String()
			mimeType := fileInput.Get("files").Call("item", 0).Get("type").String()

			log.Print("Looking at " + filename + " type " + mimeType)

			out := string(dst)
			if len(out) > 100 {
				out = out[:100] + "..."
			}

			//fileOutput.Set("innerText", out)

			r := bytes.NewReader(dst)

			ids, err := s.Identify(r, filename, mimeType)
			if err != nil {
				log.Fatal(err)
			}

			out = ""
			for _, id := range ids {
				out = out + id.String() + ":\n"
				res := s.Label(id)
				for _, value := range res {
					out = out + "    " + value[0] + ": " + value[1] + "\n"
				}
			}
			fileOutput.Set("innerText", out)

			return nil
		}))

		return nil
	}))

	fileInput.Set("disabled", false)

	<-quit
}
