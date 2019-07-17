package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/fentec-project/scenario/abe"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	keyPath := ""
	outputFile := ""
	/***** SELECT DECRYPTOR *****/
	menu :=

		`Select decryptor
	[1] Decryptor1 - {0, 2, 3}
	[2] Decryptor2 - {0, 4}
	[3] Decryptor3 - {0, 2, 5}`

	fmt.Println(menu)

	reader := bufio.NewReader(os.Stdin)

	in, _ := reader.ReadString('\n')
	selection := strings.TrimRight(in, "\r\n")

	switch selection {
	case "1":
		fmt.Println("Decryptor1 - {0, 2, 3}")
		keyPath = "keyDecryptor1.gob"
		outputFile = "decryptor1_data.xlsx"
	case "2":
		fmt.Println("Decryptor2 - {0, 4}")
		keyPath = "keyDecryptor2.gob"
		outputFile = "decryptor2_data.xlsx"
	case "3":
		fmt.Println("Decryptor3 - {0, 2, 5}")
		keyPath = "keyDecryptor3.gob"
		outputFile = "decryptor3_data.xlsx"
	default:
		fmt.Println("ERROR: Undefined decryptor")
	}
	/***** SELECT DECRYPTOR *****/

	path := "key-material/"

	var outputExcelFile *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell

	outputExcelFile = xlsx.NewFile()
	inputExcelFileName := "data/encryptor_data.xlsx"
	inputExcelFile, _ := xlsx.OpenFile(inputExcelFileName)

	for _, _sheet := range inputExcelFile.Sheets {
		sheet, _ = outputExcelFile.AddSheet("1")
		for _, _row := range _sheet.Rows {
			row = sheet.AddRow()
			i := 0
			for _, _cell := range _row.Cells {
				cell = row.AddCell()
				encoded_enc_text := _cell.String() // information to be decrypted
				//Columns whose data will be encrypted. In this case: 1, 2 and 3 (starting from 0)
				if i == 1 || i == 2 || i == 3 {
					enc_text, _ := base64.StdEncoding.DecodeString(encoded_enc_text)
					ioutil.WriteFile(path+"aux_encrypt.gob", []byte(enc_text), 0644)
					// decrypt the ciphertext with the keys of an entity
					// that has sufficient attributes
					err := abe.Decrypt(path, keyPath)
					if err != nil {
						//fmt.Printf("Failed to decrypt: %v\n", err)
						text := _cell.String()
						cell.Value = text
					} else {
						text, _ := ioutil.ReadFile(path + "aux_decrypt.gob")
						cell.Value = string(text)
					}
				} else {
					text := _cell.String()
					cell.Value = text
				}
				i++
			}
		}
		break
	}

	outputExcelFile.Save("data/" + outputFile)

	fmt.Println("\n---------- DECRYPTOR finished ----------\n")
}
