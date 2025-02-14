package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"syscall/js"
)

const apiBaseUrl = "http://localhost:8443/"
const apiKey = "oRJVT5NgdIFpflx-eZhWGSd2hvU-3cOa3qDRtVIDndc"

func uploadFile(this js.Value, p []js.Value) interface{} {
	input := p[0]
	if input.Get("files").Length() == 0 {
		return nil
	}
	file := input.Get("files").Index(0)
	fileName := file.Get("name")
	fileReader := js.Global().Get("FileReader").New()
	fileReader.Call("readAsArrayBuffer", file)

	fileReader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		arrayBuffer := fileReader.Get("result")

		// Ensure it's a Uint8Array
		uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
		data := make([]byte, uint8Array.Length())

		// Copy bytes correctly
		js.CopyBytesToGo(data, uint8Array)

		go func() {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			formFile, err := writer.CreateFormFile("file", fileName.String())
			if err != nil {
				fmt.Println("Error creating form file:", err)
				return
			}
			_, err = formFile.Write(data)
			if err != nil {
				fmt.Println("Error writing to form file:", err)
				return
			}
			err = writer.Close()
			if err != nil {
				fmt.Println("Error closing writer:", err)
				return
			}
			req, err := http.NewRequest("POST", apiBaseUrl+"files/upload/", &buf)
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}
			req.Header.Set("X-API-Key", apiKey)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Upload failed:", err)
				return
			}
			if resp.StatusCode != 200 {
				fmt.Println("Upload failed:", resp.StatusCode)
				body, _ := ioutil.ReadAll(resp.Body)
				fmt.Println("Response body:", string(body))
				return
			}
			defer resp.Body.Close()
			fmt.Println("Upload successful")
		}()
		return nil
	}))

	return nil
}

func fetchFiles(this js.Value, p []js.Value) interface{} {
	go func() {
		resp, err := http.Get(apiBaseUrl)
		if err != nil {
			fmt.Println("Error fetching files:", err)
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		var files []map[string]string
		json.Unmarshal(body, &files)
		js.Global().Get("document").Call("getElementById", "fileList").Set("innerHTML", "")
		for _, file := range files {
			item := js.Global().Get("document").Call("createElement", "li")
			item.Set("textContent", file["name"])
			js.Global().Get("document").Call("getElementById", "fileList").Call("appendChild", item)
		}
	}()
	return nil
}

func registerCallbacks() {
	js.Global().Set("uploadFile", js.FuncOf(uploadFile))
	js.Global().Set("fetchFiles", js.FuncOf(fetchFiles))
}
