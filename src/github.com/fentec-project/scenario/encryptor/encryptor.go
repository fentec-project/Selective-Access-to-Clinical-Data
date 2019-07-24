package main

import (
	"github.com/fentec-project/gofe/abe"

	"encoding/base64"
	"fmt"
	"github.com/fentec-project/scenario/adaptor"
	"github.com/tealeg/xlsx"
	"io/ioutil"
)

func main() {
	debug := true

	path := "key-material/"
	// create a msp struct out of a boolean expression representing the
	// policy specifying which attributes are needed to decrypt the ciphertext;
	// note that safety of the encryption is only proved if the mapping
	// msp.RowToAttrib from the rows of msp.Mat to attributes is injective, i.e.
	// only boolean expressions in which each attribute appears at most once
	// are allowed - if expressions with multiple appearances of an attribute
	// are needed, then this attribute can be split into more sub-attributes
	//msp, _ := abe.BooleanToMSP("((0 AND 1) OR (2 AND 3)) AND 5", false)
	msp1, _ := abe.BooleanToMSP("(0 AND 2) OR 6", false)
	msp2, _ := abe.BooleanToMSP("(0 AND 4) OR 5", false)
	msp3, _ := abe.BooleanToMSP("(0 AND 2) AND 5", false)

	var outputExcelFile *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell

	outputExcelFile = xlsx.NewFile()
	inputExcelFileName := "data/data.xlsx"
	inputExcelFile, _ := xlsx.OpenFile(inputExcelFileName)

	for _, _sheet := range inputExcelFile.Sheets {
		sheet, _ = outputExcelFile.AddSheet("1")
		for _, _row := range _sheet.Rows {
			row = sheet.AddRow()
			i := 0
			if debug {
				fmt.Println()
			}
			for _, _cell := range _row.Cells {
				var msp *abe.MSP
				if i == 1 {
					if debug {
						fmt.Printf("Column %d: ((0 AND 2) OR 6)\n", i)
					}
					msp = msp1
				} else if i == 2 {
					if debug {
						fmt.Printf("Column %d: ((0 AND 4) OR 5)\n", i)
					}
					msp = msp2
				} else if i == 3 {
					if debug {
						fmt.Printf("Column %d: ((0 AND 2) OR 6)\n", i)
					}
					msp = msp3
				} else {
					if debug {
						fmt.Printf("Column %d: without encryption\n", i)
					}
				}
				cell = row.AddCell()
				text := _cell.String() // information to be encrypted
				//Columns whose data will be encrypted. In this case: 1, 2 and 3 (starting from 0)
				if i == 1 || i == 2 || i == 3 {
					// encrypt the message msg with the decryption policy specified by the
					// msp structure
					adaptor.Encrypt(path, text, msp)
					enc_text, _ := ioutil.ReadFile(path + "aux_encrypt.gob")
					encoded_enc_text := base64.StdEncoding.EncodeToString(enc_text)
					cell.Value = encoded_enc_text
				} else {
					text := _cell.String()
					cell.Value = text
				}
				i++
			}
		}
		break
	}

	outputExcelFile.Save("data/encryptor_data.xlsx")

	fmt.Println("\n---------- ENCRYPTOR finished ----------\n")
}
