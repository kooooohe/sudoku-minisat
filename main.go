package main

import (
	"fmt"
	"math"
	"os"
)

/*
x_1_1_k x_1_2_k | x_1_3_k x_1_4_k
x_2_1_k x_2_2_k | x_2_3_k x_2_4_k
-------------
x_3_1_k x_3_2_k | x_3_3_k x_3_4_k
x_4_1_k x_4_2_k | x_4_3_k x_4_4_k

a1 a2 | b1 b2
a3 a4 | b3 b4
-------------
c1 c2 | d1 d2
c3 c4 | d3 d4
*/

const SIZE = 9

// varnum converts the (row, column, digit) to a variable number for SAT solver
func varnum(row, column, digit int) int {
	return (row-1)*SIZE*SIZE + (column-1)*SIZE + digit
}

// generateClauses generates all the clauses for the SAT solver input
func generateClauses() [][]int {
	clauses := [][]int{}

	// Each cell contains at least one number
	// a_1_1 V a_1_2 V a_1_4 V a_1_4 ...
	for row := 1; row <= SIZE; row++ {
		for column := 1; column <= SIZE; column++ {
			clause := []int{}
			for digit := 1; digit <= SIZE; digit++ {
				clause = append(clause, varnum(row, column, digit))
			}
			clauses = append(clauses, clause)
		}
	}

	// Each number appears at most once in each row
	// !a_1_1 V !a_2_1
	// !a_1_1 V !b_1_1 ...
	for row := 1; row <=SIZE; row++ {
		for digit := 1; digit <=SIZE; digit++ {
			for column1 := 1; column1 <=SIZE; column1++ {
				for column2 := column1 + 1; column2 <=SIZE; column2++ {
					clause := []int{-varnum(row, column1, digit), -varnum(row, column2, digit)}
					clauses = append(clauses, clause)
				}
			}
		}
	}

	// Each number appears at most once in each column
	// !a_1_1 V !a_3_1
	// !a_1_1 V !c_1_1 ...
	for column := 1; column <=SIZE; column++ {
		for digit := 1; digit <=SIZE; digit++ {
			for row1 := 1; row1 <=SIZE; row1++ {
				for row2 := row1 + 1; row2 <=SIZE; row2++ {
					clause := []int{-varnum(row1, column, digit), -varnum(row2, column, digit)}
					clauses = append(clauses, clause)
				}
			}
		}
	}

	// Each number appears at most once in each 2x2 subgrid
	for digit := 1; digit <=SIZE; digit++ {
		for rowStart := 1; rowStart <= SIZE; rowStart += int(math.Sqrt(SIZE)) {
			for columnStart := 1; columnStart <= SIZE; columnStart += int(math.Sqrt(SIZE)) {
				for pos := 0; pos <SIZE; pos++ {
					for pos2 := pos + 1; pos2 <SIZE; pos2++ {
						r1 := rowStart + pos/int(math.Sqrt(SIZE))
						c1 := columnStart + pos%int(math.Sqrt(SIZE))
						r2 := rowStart + pos2/int(math.Sqrt(SIZE))
						c2 := columnStart + pos2%int(math.Sqrt(SIZE))
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
	clauses := generateClauses()

	// Writing the CNF to a file
	file, err := os.Create("minisat_input.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the header
	_, err = file.WriteString(fmt.Sprintf("p cnf %d %d\n",SIZE*SIZE*SIZE, len(clauses)))
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

