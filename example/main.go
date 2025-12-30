package main

func main() {
	srv := setup()
	srv.ServeGracefully()
}
