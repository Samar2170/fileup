package main

func main() {
	registerCallbacks()
	<-make(chan struct{})
}
