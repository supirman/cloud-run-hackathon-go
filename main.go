package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	http.HandleFunc("/", handler)

	log.Printf("starting server on port :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatalf("http listen error: %v", err)
}

var directionCost = map[string]map[string]int{
	"N": {
		"N": 0,
		"E": 1,
		"S": 2,
		"W": -1,
	},
	"E": {
		"N": -1,
		"E": 0,
		"S": 1,
		"W": 2,
	},
	"S": {
		"N": 2,
		"E": -1,
		"S": 0,
		"W": 1,
	},
	"W": {
		"N": 1,
		"E": 2,
		"S": -1,
		"W": 0,
	},
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		fmt.Fprint(w, "Let the battle begin!")
		return
	}

	var v ArenaUpdate
	defer req.Body.Close()
	d := json.NewDecoder(req.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		log.Printf("WARN: failed to decode ArenaUpdate in response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := play(v)
	fmt.Fprint(w, resp)
}

func getCost(a ArenaUpdate, myId string) (playerDistance map[string]Option) {
	playerCost := make(map[string]Option)
	myState := a.Arena.State[myId]
	for key, state := range a.Arena.State {
		if key != myId {
			var nextMove = "T"
			dX := state.X - myState.X
			dY := state.Y - myState.Y

			xDir := myState.Direction
			if dX > 0 {
				xDir = "E"
			} else if dX < 0 {
				xDir = "W"
			}

			yDir := myState.Direction
			if dY > 0 {
				yDir = "S"
			} else if dY < 0 {
				yDir = "N"
			}

			if dX != 0 && dY == 0 {
				yDir = xDir
			}
			if dX == 0 && dY != 0 {
				xDir = yDir
			}

			// count X movement
			var xCost = abs(directionCost[myState.Direction][xDir]) + abs(dX) + abs(directionCost[xDir][yDir]) + max(abs(dY)-3, 0)
			// count X movement
			var yCost = abs(directionCost[myState.Direction][yDir]) + abs(dY) + abs(directionCost[yDir][xDir]) + max(abs(dX)-3, 0)

			distance := min(xCost, yCost)
			// decide next move
			if distance == 0 {
				nextMove = "T"
			} else if distance == xCost {
				if directionCost[myState.Direction][xDir] < 0 {
					nextMove = "L"
				} else if directionCost[myState.Direction][xDir] > 0 {
					nextMove = "R" // TODO: decide if need to rotate twice
				} else if dY != 0 {
					nextMove = "F"
				} else if directionCost[xDir][yDir] < 0 {
					nextMove = "L"
				} else if directionCost[xDir][yDir] > 0 {
					nextMove = "R"
				} else if dX > 3 {
					nextMove = "F"
				} else {
					nextMove = "T"
				}
			} else if distance == yCost {
				if directionCost[myState.Direction][yDir] < 0 {
					nextMove = "L"
				} else if directionCost[myState.Direction][yDir] > 0 {
					nextMove = "R" // TODO: decide if need to rotate twice
				} else if dY != 0 {
					nextMove = "F"
				} else if directionCost[yDir][xDir] < 0 {
					nextMove = "L"
				} else if directionCost[yDir][xDir] > 0 {
					nextMove = "R"
				} else if dY > 3 {
					nextMove = "F"
				} else {
					nextMove = "T"
				}
			}

			playerCost[key] = Option{Cost: distance, NextMove: nextMove}
		}
	}
	return playerCost
}

func play(input ArenaUpdate) (response string) {
	log.Printf("IN: %#v", input)

	myId := input.Links.Self.Href
	myState := input.Arena.State[myId]
	playerCost := getCost(input, myId)
	log.Printf("OPTIONS: %#v", playerCost)
	nextMove := "T"
	target := myId
	if !myState.WasHit {
		min := input.Arena.Dimensions[0] + input.Arena.Dimensions[1] + 4

		for k, option := range playerCost {
			if option.Cost < min {
				min = option.Cost
				nextMove = option.NextMove
				target = k
			}
		}

		log.Printf("Next: %#v, %#v", target, nextMove)
	} else { // run away
		min := int(^uint(0) >> 1)
		for k, state := range input.Arena.State {
			if state.Score < min && state.Score > 0 && k != myId {
				min = state.Score
				nextMove = playerCost[k].NextMove
				target = k
			}
		}
		log.Printf("RUN: %#v, %#v", target, nextMove)
	}
	return nextMove
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
