package main

import (
    "fmt";
    "bufio";
    "os";
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)

    for ;; {
	fmt.Print("Pokedex > ")
	scanner.Scan()
	words := cleanInput(scanner.Text())
	fmt.Printf("Your command was: %s\n", words[0])
    }
}
