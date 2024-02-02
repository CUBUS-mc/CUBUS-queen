package CUBUS_queen

func main(config map[string]interface{}) {
	switch config["ui_type"].(string) {
	case "c":
		cli(config)
	case "g":
		gui(config)
	}
	mainLoop()
}

func mainLoop() {
	for {
		print("> ")
	}
}
