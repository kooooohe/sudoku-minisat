package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// varnum converts the (row, column, digit) to a variable number for SAT solver
func varnum(row, column, digit int) int {
	return (row-1)*81 + (column-1)*9 + digit
}

// generateClauses generates all the clauses for the SAT solver input
func generateClauses() [][]int {
	clauses := [][]int{}

	// Each cell contains at least one number
	for row := 1; row <= 9; row++ {
		for column := 1; column <= 9; column++ {
			clause := []int{}
			for digit := 1; digit <= 9; digit++ {
				clause = append(clause, varnum(row, column, digit))
			}
			clauses = append(clauses, clause)
		}
	}

	// Each number appears at most once in each row
	for row := 1; row <= 9; row++ {
		for digit := 1; digit <= 9; digit++ {
			for column1 := 1; column1 <= 9; column1++ {
				for column2 := column1 + 1; column2 <= 9; column2++ {
					clause := []int{-varnum(row, column1, digit), -varnum(row, column2, digit)}
					clauses = append(clauses, clause)
				}
			}
		}
	}

	// Each number appears at most once in each column
	for column := 1; column <= 9; column++ {
		for digit := 1; digit <= 9; digit++ {
			for row1 := 1; row1 <= 9; row1++ {
				for row2 := row1 + 1; row2 <= 9; row2++ {
					clause := []int{-varnum(row1, column, digit), -varnum(row2, column, digit)}
					clauses = append(clauses, clause)
				}
			}
		}
	}

	// Each number appears at most once in each 3x3 subgrid
	for digit := 1; digit <= 9; digit++ {
		for rowStart := 1; rowStart <= 9; rowStart += 3 {
			for columnStart := 1; columnStart <= 9; columnStart += 3 {
				for pos := 0; pos < 9; pos++ {
					for pos2 := pos + 1; pos2 < 9; pos2++ {
						r1 := rowStart + pos/3
						c1 := columnStart + pos%3
						r2 := rowStart + pos2/3
						c2 := columnStart + pos2%3
						clause := []int{-varnum(r1, c1, digit), -varnum(r2, c2, digit)}
						clauses = append(clauses, clause)
					}
				}
			}
		}
	}

	return clauses
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudoku <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]
	baseClauses := generateClauses()

	sudoku := sudokuBoard(filename)
	clauses := toClauses(sudoku)
	clauses = append(baseClauses, clauses...)
	writeCNF(clauses)
	r := runMinisat()
	fmt.Println(r)
	if !r {
		fmt.Println("Error the file is not satisfiable")
		return
	}

	cells := randomCells()
	for _,v := range cells {
		// remove a number of cell

		// check other numbers
		sudoku[v[0]][v[1]] *= -1

		clauses := toClauses(sudoku)
		clauses = append(baseClauses, clauses...)
		writeCNF(clauses)
		r := runMinisat()
		fmt.Println(r)

		if r {
			sudoku[v[0]][v[1]] *= -1
			break
		}

		// make it blank
		sudoku[v[0]][v[1]] = 0
	}
	for _, row := range sudoku {
		fmt.Println(row)
	}
}

func randomCells() [][]int{
	rand.NewSource(time.Now().UnixNano())

	combinations := [][]int{}
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			combinations = append(combinations, []int{i, j})
		}
	}

	rand.Shuffle(len(combinations), func(i, j int) {
		combinations[i], combinations[j] = combinations[j], combinations[i]
	})
	return combinations
}

func toClauses(board [][]int) [][]int {
	clauses := [][]int{}
	for i, vs := range board {
		for j, v := range vs {
			if v == 0 {
				continue
			}
			isn := v < 0
			if isn {
				v *= -1
			}
			c := varnum(i+1, j+1, v)
			if isn {
				c *= -1
			}
			clause := []int{c}
			clauses = append(clauses, clause)
		}
	}

	return clauses
}

func sudokuBoard(fName string) [][]int {
	var sudoku [][]int
	file, err := os.Open(fName)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		numbers := strings.Split(line, ",")

		var row []int
		for _, numStr := range numbers {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				fmt.Printf("Error converting string to int: %s\n", err)
				os.Exit(1)
			}
			row = append(row, num)
		}
		sudoku = append(sudoku, row)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	for _, row := range sudoku {
		fmt.Println(row)
	}
	return sudoku
}

// overwrite
func writeCNF(clauses [][]int) {
	file, err := os.Create("create_minisat_input_9x9.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the header
	_, err = file.WriteString(fmt.Sprintf("p cnf 729 %d\n", len(clauses)))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// Write the clauses
	for _, clause := range clauses {
		for _, literal := range clause {
			_, err = file.WriteString(fmt.Sprintf("%d ", literal))
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}
		_, err = file.WriteString("0\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
}

func runMinisat() bool {
	minisatCmd := "minisat"
	minisatArgs := []string{"create_minisat_input_9x9.txt"}

	tailCmd := "tail"
	tailArgs := []string{"-n", "1"}

	cmdMinisat := exec.Command(minisatCmd, minisatArgs...)
	var outMinisat bytes.Buffer
	cmdMinisat.Stdout = &outMinisat

	err := cmdMinisat.Run()
	if err != nil {
		// fmt.Printf("Error running minisat: %s\n", err)
		// return false
	}
	// fmt.Println(outMinisat.String())

	cmdTail := exec.Command(tailCmd, tailArgs...)
	cmdTail.Stdin = &outMinisat

	var outTail bytes.Buffer
	cmdTail.Stdout = &outTail

	err = cmdTail.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return false
	}

	result := outTail.String()
	result = strings.TrimSpace(result)
	fmt.Printf("Result: %s\n", result)

	return result == "SATISFIABLE"
}
