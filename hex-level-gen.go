package main

import (
	"encoding/json"
	"errors"
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

	maxClusterProb = .7
	minClusterProb = .2

	outDir = "levels"

	targetMultiple = 50

	minTargetScore = 100
	maxTargetScore = 1500

	maxPushInterval = 15
	minPushInterval = 5

	maxNiceness = .5
	minNiceness = .2

	maxBombProb = .33
	minBombProb = .2

	maxSuperProb = .075
	minSuperProb = .01
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

	hexProb := 1.0 - maxBombProb - maxSuperProb
	bombProb := maxBombProb
	superProb := maxSuperProb

	var currLevel int
	for currLevel = 0; currLevel < numLevels; currLevel++ {
		currJsonMap := make(map[string]interface{})

		// Linear decrease in cluster probability as level number increases
		//clusterProb := maxClusterProb - (maxClusterProb-minClusterProb)*(float64(currLevel)/float64(numLevels))
		clusterProb := linearScale(0, numLevels, minClusterProb, maxClusterProb, numLevels-currLevel)

		// Linear decrease in niceness probability as level number increases
		//niceness := maxNiceness - (maxNiceness-minNiceness)*(float64(currLevel)/float64(numLevels))
		niceness := linearScale(0, numLevels, minNiceness, maxNiceness, numLevels-currLevel)

		bombProb = linearScale(0, numLevels, minBombProb, maxBombProb, numLevels-currLevel)
		superProb = linearScale(0, numLevels, minSuperProb, maxSuperProb, numLevels-currLevel)
		hexProb = 1.0 - bombProb - superProb

		pushInterval := linearScale(0, numLevels, minPushInterval, maxPushInterval, numLevels-currLevel)
		pushInterval = math.Ceil(pushInterval)

		targetScore := linearScale(0, numLevels, minTargetScore, maxTargetScore, currLevel)

		currJsonMap["target"] = nearestMultiple(int(targetScore), targetMultiple)
		currJsonMap["pushInterval"] = int(pushInterval)
		currJsonMap["hexProb"] = hexProb
		currJsonMap["bombProb"] = bombProb
		currJsonMap["superProb"] = superProb
		currJsonMap["niceness"] = niceness

		/*
			bombProb *= bombDecreaseFactor
			bombProb = math.Max(minBombProb, bombProb)
			superProb *= superDecreaseFactor
			superProb = math.Max(minSuperProb, superProb)
			hexProb = 1.0 - bombProb - superProb
		*/
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

func linearScale(fromMin, fromMax, toMin, toMax, fromInput interface{}) float64 {
	var err error
	fromMinFloat, err := toFloat64(fromMin)
	if !checkErr(err) {
		return -1.0
	}
	fromMaxFloat, err := toFloat64(fromMax)
	if !checkErr(err) {
		return -1.0
	}
	toMinFloat, err := toFloat64(toMin)
	if !checkErr(err) {
		return -1.0
	}
	toMaxFloat, err := toFloat64(toMax)
	if !checkErr(err) {
		return -1.0
	}
	fromInputFloat, err := toFloat64(fromInput)
	if !checkErr(err) {
		return -1.0
	}

	return (toMaxFloat-toMinFloat)*((fromInputFloat-fromMinFloat)/(fromMaxFloat-fromMinFloat)) + toMinFloat
}

func toFloat64(x interface{}) (float64, error) {
	switch x := x.(type) {
	case uint8:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case int:
		return float64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return float64(x), nil
	default:
		return math.NaN(), errors.New("Cannot convert to float - value has unknown type.")
	}
}

func nearestMultiple(num, multiple int) int {
	return (num + multiple - 1) / multiple * multiple
}

func checkErr(err error) bool {
	if err != nil {
		log.Println(err.Error())
		return false
	} else {
		return true
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
