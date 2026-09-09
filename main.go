package main

import (
    "bufio";
    "encoding/json";
    "fmt";
    "io";
    "log";
    "net/http";
    "os";
)

type config struct {
    cmdRegistry	map[string]cliCommand
    mapNext	string
    mapPrev	string
}

type locationAreasResult struct {
    Count	int		`json:"count"`
    Next	string		`json:"next"`
    Previous	string		`json:"previous"`
    Results	[]locationArea	`json:"results"`
}

type locationArea struct {
    Name	string	`json:"name"`
    Url		string	`json:"url"`
}

type cliCommand struct {
    name 	string
    description	string
    callback 	func(*config) error
}
var commands map[string]cliCommand

func getPokeEndPoint(path string) string {
    prefix := "https://pokeapi.co/api/v2/"
    return fmt.Sprintf("%s%s", prefix, path)
}

func commandMapNext(state *config) error {
    if state.mapNext == "" {
	fmt.Printf("you're on the last page\n")
    } else {
	cmdMap(state.mapNext, state)
    }
    return nil
}

func commandMapPrev(state *config) error {
    if state.mapPrev == "" {
	fmt.Printf("you're on the first page\n")
    } else {
	cmdMap(state.mapPrev, state)
    }
    return nil
}

func cmdMap(url string, state *config) error {
    res, err := http.Get(url)
    if err != nil {
	log.Fatalf("Error reaching location areas: %v", err)
    }
    defer res.Body.Close()

    data, err := io.ReadAll(res.Body)
    if err != nil || res.StatusCode > 299 {
	log.Fatalf("Error returned from location areas: %v", err)
    }
    //    fmt.Printf("%s\n", data)
    
    locAreas := locationAreasResult{}
    if err := json.Unmarshal(data, &locAreas); err != nil {
	log.Fatalf("Error parsing location area data: %v", err)
    }

    state.mapNext = locAreas.Next
    state.mapPrev = locAreas.Previous

    fmt.Printf("\n")
    for _, area := range locAreas.Results {
	fmt.Printf("%s\n", area.Name)
    }
    fmt.Printf("\n")

    return nil
}

func commandExit(*config) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp(*config) error {
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
	    name: 		"exit",
	    description:	"Exit the Pokedex",
	    callback: 		commandExit,
        },
	"help": {
	    name: 		"help",
	    description:	"Displays a help message",
	    callback:		commandHelp,
        },
	"map": {
	    name:		"map",
	    description: 	"Displays [first/next] 20 location area names",
	    callback: 		commandMapNext,
	},
	"mapb": {
	    name:		"mapb",
	    description: 	"Displays previous 20 location area names",
	    callback: 		commandMapPrev,
	},
    }
    stateConfig := config {
	cmdRegistry:	commands,
	mapNext:	getPokeEndPoint("location-area"),
	mapPrev:	"",
    }

    scanner := bufio.NewScanner(os.Stdin)

    for ;; {
	fmt.Print("Pokedex > ")
	scanner.Scan()
	words := cleanInput(scanner.Text())

	if _, ok := commands[words[0]]; !ok {
	    fmt.Println("Unknown command\n")
	} else {
		commands[words[0]].callback(&stateConfig)
	}
    }
}
