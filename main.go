package main

import (
    "bufio";
    "encoding/json";
    "fmt";
    "io";
    "log";
	"math";
    "net/http";
	"net/url";
    "os";
	"strconv";
    "time";

    pCache "github.com/farhanjiwani/internal/pokecache";
)

type config struct {
    cmdRegistry	map[string]cliCommand
    mapNext		string
    mapPrev		string
    cache		pCache.Cache
}

type locationAreasResult struct {
    Count		int				`json:"count"`
    Next		string			`json:"next"`
    Previous	string			`json:"previous"`
    Results		[]locationArea	`json:"results"`
}

type locationArea struct {
    Name	string	`json:"name"`
    Url		string	`json:"url"`
}

type locationPokemonResult struct {
	Name		string		`json:"name"`
	Encounters	[]encounter `json:"pokemon_encounters"`
}

type encounter struct {
	Pokemon		struct{
		Name	string	`json:"name"`
	}	`json:"pokemon"`
}

type cliCommand struct {
    name 		string
    description	string
    callback 	func(*config, string) error
}
var commands map[string]cliCommand

/** Parses endpoint URL for the offset parameter **/
func getOffsetParam(urlString string) string {
	rawURL, err := url.Parse(urlString);
	if err != nil {
		log.Fatalf("Error parsing endpoint for cache: %v", err)
	}

	query := rawURL.Query()
	offset := query.Get("offset")
	if offset == "" {
		offset = "0"
	}
	return offset
}

func getPokeEndPoint(path string) string {
    prefix := "https://pokeapi.co/api/v2/"
    return fmt.Sprintf("%s%s", prefix, path)
}

/**
 * Stores page content and their pagination links.
 * Uses url to parse offset value to use as cache key.
 */
func addAreasToCache(state *config, url, value, previous, next string) {
	offset := getOffsetParam(url)
	state.cache.Add(offset, []byte(value))
	state.cache.Add(offset + "p", []byte(previous))
	state.cache.Add(offset + "n", []byte(next))
}

func getLocationAreas(state *config, url string) {
	offset := getOffsetParam(url)
	page, _ := strconv.Atoi(offset)
	page = int(math.Floor(float64(page) / 20) + 1)
	listSrc := ""

	values, exists := state.cache.Get(offset)
	if exists {
		fmt.Printf("%s", values)
		prev, _ := state.cache.Get(offset + "p")
		next, _ := state.cache.Get(offset + "n")
		state.mapPrev = string(prev)
		state.mapNext = string(next)
		listSrc = " (cached)"

	} else {
		cmdMap(url, state)
    }
	fmt.Printf("[Page %d%s]\n", page, listSrc)
}

func getAreaEncounters(state *config, locationArea string) {
	listSrc := ""
	values, exists := state.cache.Get(locationArea)
	if exists {
		fmt.Printf("%s", values)
		listSrc = "\n (cached)\n"
	} else {
		cmdExplore(locationArea, state)
	}
	fmt.Printf("%s\n", listSrc)
}

func commandMapNext(state *config, _nil string) error {
    if state.mapNext == "" {
		fmt.Printf("you're on the last page\n")
	} else {
		getLocationAreas(state, state.mapNext)
	}
	return nil

}

func commandMapPrev(state *config, _ string) error {
    if state.mapPrev == "" {
		fmt.Printf("you're on the first page\n")
    } else {
		getLocationAreas(state, state.mapPrev)
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
    
    locAreas := locationAreasResult{}
    if err := json.Unmarshal(data, &locAreas); err != nil {
		log.Fatalf("Error parsing location area data: %v", err)
    }

	// Format, cache and output
	locationsFmt := "\n"
    for _, area := range locAreas.Results {
		locationsFmt += fmt.Sprintf("%s\n", area.Name)
    }
	locationsFmt += "\n"
	addAreasToCache(state, url, locationsFmt, locAreas.Previous, locAreas.Next)
	fmt.Printf("%s", locationsFmt)

	// Update pagination
    state.mapPrev = locAreas.Previous
    state.mapNext = locAreas.Next

    return nil
}

func cmdExplore(location string, state *config) error {
	endpoint := getPokeEndPoint("location-area/" + location)

    res, err := http.Get(endpoint)
    if err != nil {
		log.Fatalf("Error reaching location area '%s': %v", location, err)
    }
    defer res.Body.Close()

    data, err := io.ReadAll(res.Body)
    if err != nil || res.StatusCode > 299 {
		log.Fatalf("Error returned from location area '%s': %v", location, err)
    }

	encounters := locationPokemonResult{}
    if err := json.Unmarshal(data, &encounters); err != nil {
		log.Fatalf("Error parsing '%s' encounter data: %v", location, err)
    }
	fmt.Printf("Exploring %s...\nFound Pokemon:\n", encounters.Name)

	pokeList := ""
	for _, p := range encounters.Encounters {
		pokeList += fmt.Sprintf(" - %s\n", p.Pokemon.Name)
	}
	fmt.Println(pokeList)

	state.cache.Add(location, []byte(pokeList))
	return nil
}

func commandExplore(state *config, location string) error {
	if len(location) == 0 {
		fmt.Println("LOCATION REQUIRED! Use `map` or `mapb` for location names")
	}
	
	getAreaEncounters(state, location)
	return nil
}

func commandExit(*config, string) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp(*config, string) error {
    fmt.Println("Welcome to the Pokedex!")
    fmt.Println("Usage:")

    for _, command := range commands {
	fmt.Printf("%s:\t%s\n", command.name, command.description)
    }
    fmt.Printf("\n")

    return nil
}

func main() {
	pc := pCache.NewCache(5 * time.Second)
    commands = map[string]cliCommand{
		"exit": {
			name:	 		"exit",
			description:	"Exit the Pokedex",
			callback: 		commandExit,
		},
		"help": {
			name: 			"help",
			description:	"Displays a help message",
			callback:		commandHelp,
		},
		"map": {
			name:			"map",
			description: 	"Displays [first/next] 20 location area names",
			callback: 		commandMapNext,
		},
		"mapb": {
			name:			"mapb",
			description: 	"Displays previous 20 location area names",
			callback: 		commandMapPrev,
		},
		"explore": {
			name:			"explore",
			description: 	"List Pokemon at desired location",
			callback: 		commandExplore,
		},
    }
    stateConfig := config {
		cmdRegistry:	commands,
		mapNext:	getPokeEndPoint("location-area"),
		mapPrev:	"",
		cache:		pc,
    }

    scanner := bufio.NewScanner(os.Stdin)

    for ;; {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		arg := ""
		if len(words) > 1 {
			arg = words[1]
		}

		if _, ok := commands[words[0]]; !ok {
			fmt.Println("Unknown command")
		} else {
			commands[words[0]].callback(&stateConfig, arg)
		}
    }
}
