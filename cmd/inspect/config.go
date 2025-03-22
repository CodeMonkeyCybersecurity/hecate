// runInspectConfig presents an interactive menu for inspection
func runInspectConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Inspect Configurations ===")
	fmt.Println("Select the resource you want to inspect:")
	fmt.Println("1) Inspect Certificates")
	fmt.Println("2) Inspect docker-compose file")
	fmt.Println("3) Inspect Eos backend web apps configuration")
	fmt.Println("4) Inspect Nginx defaults")
	fmt.Println("5) Inspect all configurations")
	fmt.Print("Enter choice (1-5): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		inspectCertificates()
	case "2":
		inspectDockerCompose()
	case "3":
		inspectEosConfig()
	case "4":
		inspectNginxDefaults()
	case "5":
		inspectCertificates()
		inspectDockerCompose()
		inspectEosConfig()
		inspectNginxDefaults()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}
