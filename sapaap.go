package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"encoding/csv"
	"log"
	"os"
	"strings"
	"unicode/utf16"
	//"github.com/sergi/go-diff/diffmatchpatch"
)

var err error

func main() {
	// set commad arguments description
	flag.Usage = customCmdHelp

	// parse arguments and set variables
	endianF := flag.Bool("BE", false, "set if audit file from BigEndian system (AIX, HP-UX, Solaris, etc)")
	nucF := flag.Bool("NUC", false, "set if SAP system is NON-UNICODE")
	delimiterF := flag.String("d", ",", "delimiter to separate values in output records (CSV)")
	appendStringF := flag.String("a", "", "string to append to every row in result data\nexample: \"$HOST,$SAPSYSTEM\" ")
	printFormatHelpF := flag.Bool("describe", false, "get audit file format description")
	flag.Parse()

	if *printFormatHelpF {
		printFormatHelp()
		return
	}

	// buffer size UC/NUC
	var buflen int = 400
	if *nucF {
		buflen = 200
	}

	// check and set input
	inf := os.Stdin
	if flag.Arg(0) != "" {
		inf, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Printf("ERROR: Can't open file: %s\n", flag.Arg(0))
			return
		}
		defer inf.Close()
	}

	var buffer []byte
	i := 1
	reader := bufio.NewReader(inf)

	// Verificamos si existe el archivo logs.csv
	if _, err := os.Stat("./primer_logs.csv"); os.IsNotExist(err) {

		// Si no existe, entonces se crea
		f, err:= os.Create("primer_logs.csv")
		if err != nil {
			log.Fatal(err)
		}
		for {
			// clean buffer and read record by bytes
			buffer = nil
			j := 0
			for {
				if j >= buflen {
					break
				}
				ibyte, err := reader.ReadByte()
				if err != nil && err != io.EOF {
					panic(err)
				} else if err == io.EOF {
					if j != 0 {
						log.Printf("ERROR: The last string has wrong byte length (not %v): %v\n", buflen, j)
					}
					return
				}
				buffer = append(buffer, ibyte)
				j++
			}
	
			//check record lenght
			if len(buffer) != buflen {
				log.Printf("ERROR: The string #%v has wrong byte length (not %v): %v\n", i, buflen, len(buffer))
				return
			}
			// decode to utf-8 runes
			runes, err := DecodeUtf16(buffer, *endianF)
			if err != nil {
				log.Printf("ERROR: Can't decode string #%v\n%v", i, err)
			}
			// parse and convert to hive csv
			record, err := parseAndConvert(runes, *delimiterF)
			if err != nil {
				log.Printf("ERROR: Can't parse and convert string #%v %q\n%v", i, record, err)
			}
			// append string
			if *appendStringF != "" {
				record += "," + strings.Trim(*appendStringF, "\"")
			}
	
			// string result output
			fmt.Println("Primeros logs ingresados")

			_, err2:= f.WriteString(record + "\n")
	
			if err2 != nil {
			  log.Fatal(err2)
			}

			//inc string number
			i++
		}
		defer f.Close()
	}else{
		f2, err2:= os.Create("segundo_logs.csv")
		if err2 != nil {
			log.Fatal(err)
		}
		
		for {
			// clean buffer and read record by bytes
			buffer = nil
			j := 0
			for {
				if j >= buflen {
					break
				}
				ibyte, err := reader.ReadByte()
				if err != nil && err != io.EOF {
					panic(err)
				} else if err == io.EOF {
					if j != 0 {
						log.Printf("ERROR: The last string has wrong byte length (not %v): %v\n", buflen, j)
					}
					return
				}
				buffer = append(buffer, ibyte)
				j++
			}
	
			//check record lenght
			if len(buffer) != buflen {
				log.Printf("ERROR: The string #%v has wrong byte length (not %v): %v\n", i, buflen, len(buffer))
				return
			}
			// decode to utf-8 runes
			runes, err := DecodeUtf16(buffer, *endianF)
			if err != nil {
				log.Printf("ERROR: Can't decode string #%v\n%v", i, err)
			}
			// parse and convert to hive csv
			record, err := parseAndConvert(runes, *delimiterF)
			if err != nil {
				log.Printf("ERROR: Can't parse and convert string #%v %q\n%v", i, record, err)
			}
			// append string
			if *appendStringF != "" {
				record += "," + strings.Trim(*appendStringF, "\"")
			}
	
			// string result output
			
			// fmt.Println("Segundos logs ingresados")

			_, err3:= f2.WriteString(record + "\n")
	
			if err3 != nil {
			  log.Fatal(err2)
			}

			//inc string number
			i++
		}
		defer f2.Close()

		// primer_logs,err4 := os.Open("primer_logs.csv")
		// if err4 != nil{
		// 	fmt.Println(err4)
		// }
		
		// b,err5 := ioutil.ReadAll(primer_logs)
		// if err5 != nil{
		// 	fmt.Println(err5)
		// }

		// prueba, err6:= os.Create("prueba.csv")
		// if err6 != nil {
		// 	log.Fatal(err)
		// }
		
		// prueba.WriteString(string(b))
	}

}




func customCmdHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Use %s [options] <filename>\nor in pipe <output> | %s [options]\n", os.Args[0], os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(flag.CommandLine.Output(), "\n")
}

func printFormatHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), " Audit file format:                                                             \n")
	fmt.Fprintf(flag.CommandLine.Output(), " Pos   Size    Field name      Description                                      \n")
	fmt.Fprintf(flag.CommandLine.Output(), " ----------------------------------------                                       \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 1     [1]     Version         file format version (I guess)                    \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 2     [3]     MessageID       audit message identificator                      \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 5     [4]     Year                                                             \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 9     [2]     Month                                                            \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 11    [2]     Day                                                              \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 13    [2]     Hour                                                             \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 15    [2]     Minute                                                           \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 17    [2]     Second                                                           \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 19    [7]     OS_PID          OS Process ID                                    \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 26    [5]     SAP_PID         SAP Process ID                                   \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 31    [1]     LogonType       Type of connection (Dialog, RFC, etc)            \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 32    [1]     SAP_PID_hex     SAP Process ID in hex (I guess)                  \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 33    [8]     Unknown         I couldn't detect what that field means          \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 41    [12]    UserName        User LOGIN                                       \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 53    [20]    Transaction                                                      \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 73    [40]    Report                                                           \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 113   [3]     Mandant                                                          \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 116   [1]     SessionID                                                        \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 117   [64]    Parameters      most of messages has parameters, so here they are\n")
	fmt.Fprintf(flag.CommandLine.Output(), " 181   [20]    Terminal        host name user's PC                              \n")
	fmt.Fprintf(flag.CommandLine.Output(), "																				\n")
	fmt.Fprintf(flag.CommandLine.Output(), " The record length is 200 char symbols.                                         \n")
	fmt.Fprintf(flag.CommandLine.Output(), " And 200 bytes in NUC system.		                                            \n")
	fmt.Fprintf(flag.CommandLine.Output(), " In case of UINCODE system - 400 bytes because of 2 bytes per 1 symbol          \n")
	fmt.Fprintf(flag.CommandLine.Output(), "																				\n")
	fmt.Fprintf(flag.CommandLine.Output(), "    																			\n")
	fmt.Fprintf(flag.CommandLine.Output(), " Output CSV format:                                                             \n")
	fmt.Fprintf(flag.CommandLine.Output(), " Pos   Size    Field name      Description                                      \n")
	fmt.Fprintf(flag.CommandLine.Output(), " ----------------------------------------                                       \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 1     [10]    Date            YYYY-MM-DD ISO8601                               \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 2     [19]    Timestamp       YYYY-MM-DD HH:MM:ss                              \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 3     [3]     MessageID                                                        \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 4     [1]     LogonType                                                        \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 5     [0-12]  UserName                                                         \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 6     [0-20]  Transaction                                                      \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 7     [7]     OS_PID          OS Process ID                                    \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 8     [5]     SAP_PID         SAP Process ID                                   \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 9     [0-40]  Report                                                           \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 10    [3]     Mandant                                                          \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 11    [1]     SessionID                                                        \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 12    [0-64]  Parameters      most of messages has parameters, so here they are\n")
	fmt.Fprintf(flag.CommandLine.Output(), " 13    [0-20]  Terminal        host name user's PC                              \n")
	fmt.Fprintf(flag.CommandLine.Output(), " 14... [...]   <append>        custom appended string                           \n")

}

func DecodeUtf16(b []byte, endianF bool) ([]rune, error) {
	ints := make([]uint16, len(b)/2)
	if endianF {
		err := binary.Read(bytes.NewReader(b), binary.BigEndian, &ints)
		if err != nil {
			return nil, err
		}
	} else {
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &ints)
		if err != nil {
			return nil, err
		}
	}
	return utf16.Decode(ints), nil
}

func parseAndConvert(runes []rune, d string) (string, error) {
	// remove:			version, sap_pid_hex and unknown.
	// add/convert:		date as hive date, date&time as hive timestamp
	// trim:			username, transaction, report, parameters and terminal

	var result []string

	// record["version"] 	 = string(runes[0])
	// record["messageid"]   = string(runes[1:4])
	// record["year"] 		 = string(runes[4:8])
	// record["month"] 	     = string(runes[8:10])
	// record["day"]         = string(runes[10:12])
	// record["hour"]        = string(runes[12:14])
	// record["minute"]      = string(runes[14:16])
	// record["second"]      = string(runes[16:18])
	// record["os_pid"]      = string(runes[18:25])
	// record["sap_pid"]     = string(runes[25:30])
	// record["logontype"]   = string(runes[30])
	// record["sap_pid_hex"] = string(runes[31])
	// record["unknown"]     = string(runes[32:40])
	// record["username"]    = string(runes[40:52])
	// record["transaction"] = string(runes[52:72])
	// record["report"]      = string(runes[72:112])
	// record["mandant"]     = string(runes[112:115])
	// record["sessionid"]   = string(runes[115])
	// record["parameters"]  = string(runes[116:180])
	// record["terminal"]    = string(runes[180:])

	// date
	result = append(result, string(runes[4:8])+"-"+string(runes[8:10])+"-"+string(runes[10:12]))
	// timestamp
	result = append(result, result[0]+" "+string(runes[12:14])+":"+string(runes[14:16])+":"+string(runes[16:18]))
	// message id
	result = append(result, string(runes[1:4]))
	// logon type
	result = append(result, string(runes[30]))
	// username
	result = append(result, strings.Trim(string(runes[40:52]), " "))
	// transaction
	result = append(result, strings.Trim(string(runes[52:72]), " "))
	// os_pid
	result = append(result, string(runes[18:25]))
	// sap_pid
	result = append(result, string(runes[25:30]))
	// report
	result = append(result, strings.Trim(string(runes[72:112]), " "))
	// mndt
	result = append(result, string(runes[112:115]))
	// session_id
	result = append(result, string(runes[115]))
	// parameters
	result = append(result, strings.Trim(string(runes[116:180]), " "))
	// terminal
	result = append(result, strings.Trim(string(runes[180:]), " "))

	return strings.Join(result, d), nil
}
