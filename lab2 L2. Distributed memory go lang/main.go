package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Field struct {
	Name   string
	Number int
	GPA    float64
}

var count = 0

func compute(input Field) bool {
	// Check if the number is >= 200 and GPA is >= 6
	return input.Number >= 200 && input.GPA >= 6
}

// Read
func ReadFromFile(filename string) []Field {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var fields []Field
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 3 {
			name := parts[0]
			number, _ := strconv.Atoi(parts[1])
			gpa, _ := strconv.ParseFloat(parts[2], 64)
			field := Field{Name: name, Number: number, GPA: gpa}
			fields = append(fields, field)
		}
	}
	return fields
}

// Write
func WriteToFile(filename string, resultFields []Field) {
	resultFile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating result file:", err)
		os.Exit(1)
	}
	defer resultFile.Close()

	if len(resultFields) > 0 {
		resultFile.WriteString("-----------------------------------------------------\n")
		resultFile.WriteString("|    Name         |     Credits   |   GPA        |\n")
		resultFile.WriteString("-----------------------------------------------------\n")
		for _, field := range resultFields {
			if compute(field) {
				resultFile.WriteString(fmt.Sprintf("| %-15s | %12d | %12.2f |\n", field.Name, field.Number, field.GPA))
			}
		}
		resultFile.WriteString("-----------------------------------------------------\n")
		resultFile.WriteString(fmt.Sprintf("Number of data : %d\n", len(resultFields)))

	} else {
		resultFile.WriteString("No data is filtered\n")
	}
}

func DataProcess(dataChannel chan Field, resultChannel chan Field, ask chan bool, terminationChannel chan bool) {
	dataArray := make([]Field, 0, 13)
	defer close(resultChannel)
	defer close(terminationChannel)
	for {
		select {
		case field := <-dataChannel:
			dataArray = append(dataArray, field)
		case <-ask:
			if len(dataArray) > 0 {
				fieldToSend := dataArray[0]
				if fieldToSend.Name == "NULL" {
					for i := 0; i < 3; i++ {
						terminationChannel <- true
					}
					fmt.Println("Data process terminated")
					return
				}
				dataArray = dataArray[1:]
				resultChannel <- fieldToSend
			}

		}
	}
}

func Worker(wr chan Field, resultChannel chan Field, ask chan bool, terminationChannel chan bool) {

	for {
		select {
		//case request <- "request":
		case ask <- true:

		case field, ok := <-resultChannel:

			if ok {
				if compute(field) {
					wr <- field
				}
			}

		case <-terminationChannel:
			fmt.Println("Worker terminated")
			count++
			if count == 3 {
				close(wr)
			}

			return
		}
	}
}

func ResultProcess(wr chan Field, rm chan []Field) {
	defer close(rm)
	Res := make([]Field, 0, 50)
	for field := range wr {
		i := sort.Search(len(Res), func(i int) bool {
			return Res[i].Name >= field.Name
		})
		if i == len(Res) || Res[i].Name != field.Name {
			Res = append(Res, Field{})
			copy(Res[i+1:], Res[i:])
			Res[i] = field
		}
	}
	rm <- Res

	fmt.Println("Result finished")
}

// Main function
func main() {
	resultFile := "output.txt"
	dataFile2 := "f3.txt"
	var fields = ReadFromFile(dataFile2)
	workerCount := 3
	datachannel := make(chan Field) // main to data process
	rc := make(chan Field)          // result channel
	wr := make(chan Field)          // worker to result
	rm := make(chan []Field)        // result to main
	terminationChannel := make(chan bool, 3)
	var Dummy = Field{"NULL", 0, 0.0}

	ComputedArray := make([]Field, 0)
	ask := make(chan bool)

	for i := 0; i < workerCount; i++ {
		go Worker(wr, rc, ask, terminationChannel)
	}
	go DataProcess(datachannel, rc, ask, terminationChannel)

	// adding one by one
	for _, field := range fields {
		datachannel <- field
	}
	datachannel <- Dummy
	close(datachannel)

	go ResultProcess(wr, rm)
	for field := range rm {
		ComputedArray = field
	}

	WriteToFile(resultFile, ComputedArray)
}
