package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	numRows   = 20
	numCols   = 9
	numLevels = 250

	minClusterProb = .2
	maxClusterProb = .75

	outDir = "levels"

	bombDecreaseFactor  = .995
	superDecreaseFactor = .97

	minNiceness = .2
	maxNiceness = .5

	minBombProb  = .15
	minSuperProb = .025
)

var (
	colors = []string{"blue", "gray", "green", "pink", "red", "yellow"}

	commonAdjIndices = [][]int{
		{0, -1},
		{0, +1},
		{-1, 0},
		{+1, 0},
	}

	evenAdjIndices = [][]int{
		{+1, -1},
		{+1, +1},
	}

	oddAdjIndices = [][]int{
		{-1, -1},
		{-1, +1},
	}
)

func main() {
	rand.Seed(time.Now().Unix())

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.Mkdir(outDir, 0644)
	}

	hexProb := .65
	bombProb := .25
	superProb := .1

	var currLevel int
	for currLevel = 0; currLevel < numLevels; currLevel++ {
		currJsonMap := make(map[string]interface{})

		// Linear decrease in cluster probability as level number increases
		clusterProb := maxClusterProb - (maxClusterProb-minClusterProb)*(float64(currLevel)/float64(numLevels))

		// Linear decrease in niceness probability as level number increases
		niceness := maxNiceness - (maxNiceness-minNiceness)*(float64(currLevel)/float64(numLevels))

		currJsonMap["target"] = 100 + (currLevel/10)*50
		currJsonMap["pushInterval"] = math.Max(5, (float64)(10-(currLevel/25)))
		currJsonMap["hexProb"] = hexProb
		currJsonMap["bombProb"] = bombProb
		currJsonMap["superProb"] = superProb
		currJsonMap["niceness"] = niceness

		bombProb *= bombDecreaseFactor
		bombProb = math.Max(minBombProb, bombProb)
		superProb *= superDecreaseFactor
		superProb = math.Max(minSuperProb, superProb)
		hexProb = 1.0 - bombProb - superProb

		var currRow, currCol int
		for currRow = 1; currRow <= numRows; currRow++ {

			currRowStr := getRowString(currRow)
			currJsonMap[currRowStr] = make([]string, numCols)

			for currCol = 0; currCol < numCols; currCol++ {
				randVal := rand.Float64()
				if randVal < clusterProb {
					adjacentModeColor := getAdjacentModeColor(currJsonMap, currRow, currCol)
					if adjacentModeColor == "" {
						currJsonMap[currRowStr].([]string)[currCol] = colors[rand.Intn(len(colors))]
					} else {
						currJsonMap[currRowStr].([]string)[currCol] = adjacentModeColor
					}
				} else {
					currJsonMap[currRowStr].([]string)[currCol] = colors[rand.Intn(len(colors))]
				}
			}
			//currJsonMap[currRowStr] = []string{"blue", "gray", "green", "pink", "red", "yellow", "blue", "gray", "green"}
		}

		outJsonData, err := json.MarshalIndent(currJsonMap, "", "    ")
		if err != nil {
			log.Fatal("Error marshalling final level data for output: ", err)
		}

		err = ioutil.WriteFile(filepath.Join(outDir, strconv.Itoa(currLevel+1)+".json"), outJsonData, 0644)
		if err != nil {
			log.Fatal("Error writing final json data to level file: ", err)
		}
	}
}

func getAdjacentModeColor(currJsonMap map[string]interface{}, row, col int) string {
	colorCounts := make(map[string]int)

	for _, indices := range commonAdjIndices {
		adjRow := row + indices[0]
		adjCol := col + indices[1]

		if indicesValid(adjRow, adjCol) {
			if row, ok := currJsonMap[getRowString(adjRow)]; ok {
				color := row.([]string)[adjCol]
				colorCounts[color] = colorCounts[color] + 1
			}
		}
	}

	if col%2 == 0 {
		for _, indices := range evenAdjIndices {
			adjRow := row + indices[0]
			adjCol := col + indices[1]

			if indicesValid(adjRow, adjCol) {
				if row, ok := currJsonMap[getRowString(adjRow)]; ok {
					color := row.([]string)[adjCol]
					colorCounts[color] = colorCounts[color] + 1
				}
			}
		}
	} else {
		for _, indices := range oddAdjIndices {
			adjRow := row + indices[0]
			adjCol := col + indices[1]

			if indicesValid(adjRow, adjCol) {
				if row, ok := currJsonMap[getRowString(adjRow)]; ok {
					color := row.([]string)[adjCol]
					colorCounts[color] = colorCounts[color] + 1
				}
			}
		}
	}

	var modeColors []string
	modeColorCount := 0
	for color, count := range colorCounts {
		if count > modeColorCount {
			modeColors = []string{color}
			modeColorCount = count
		} else if count == modeColorCount {
			modeColors = append(modeColors, color)
		}
	}

	if len(modeColors) == 0 {
		return ""
	} else {
		return modeColors[rand.Intn(len(modeColors))]
	}
}

func indicesValid(row int, col int) bool {
	return row > 0 && row <= numRows && col >= 0 && col < numCols
}

func getRowString(row int) string {
	return "row" + strconv.Itoa(row)
}
