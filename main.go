package main

import (
    "fmt";
    "bufio";
    "os";
)

type cliCommand struct {
    name string
    description string
    callback func() error
}

var commands map[string]cliCommand

func commandExit() error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp() error {
    fmt.Println("Welcome to the Pokedex!")
    fmt.Println("Usage:\n")

    for _, command := range commands {
	fmt.Printf("%s:\t%s\n", command.name, command.description)
    }
    fmt.Printf("\n")

    return nil
}

func main() {
    commands = map[string]cliCommand{
	"exit": {
	    name: "exit",
	    description: "Exit the Pokedex",
	    callback: commandExit,
        },
	"help": {
	    name: "help",
	    description: "Displays a help message",
	    callback: commandHelp,
        },
    }

   scanner := bufio.NewScanner(os.Stdin)

    for ;; {
	fmt.Print("Pokedex > ")
	scanner.Scan()
	words := cleanInput(scanner.Text())

	if _, ok := commands[words[0]]; !ok {
	    fmt.Println("Unknown command\n")
	} else {
	    commands[words[0]].callback()
	}
    }
}
