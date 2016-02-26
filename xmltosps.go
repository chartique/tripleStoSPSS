/*

This program converts Triple-S XML files in to an SPS-syntax file to be able to import the data
to an SPSS Statistics file.

CREATOR: HEKTOR SUHR, 2016

Example use:

$ xmltosps C:/MySurvey.xml C:/MySurvey.asc

Will result in an MySurvey.sps file to be created in the same folder as the executable xmltosps.exe

*/


package main
import (
	"encoding/xml"
	"fmt"
	"os"
	"io/ioutil"
	"path"
	"strings"
	"log"
)


/* Structures the Triple-S format */
type Variables struct {
	XMLName		xml.Name		`xml:"sss"`
	Variable	[]Variable		`xml:"survey>record>variable"`
}

type Variable struct {
	XMLName		xml.Name		`xml:"variable"`
	Type		string			`xml:"type,attr"`
	Name		string			`xml:"name"`
	Label		string			`xml:"label"`
	Position	Posit
	Vals		[]Val			`xml:"values>value"`
}

type Posit struct {
	XMLName		xml.Name		`xml:"position"`
	Start		int			`xml:"start,attr"`
	Finish		int			`xml:"finish,attr"`
}

type Val struct {
	Value		int			`xml:"code,attr"`
	Name		string			`xml:",chardata"`
}


/* Helps determine what kind of a variable it is and appends the correct extension to the DATA LIST */
func (v Variable) VarType() string {
	if v.Type == "character" || v.Type == "time" {
		return fmt.Sprintf(" (A)")
	} else if v.Type == "date"{
		return fmt.Sprintf(" (A)")
	} else {
		return fmt.Sprintf("")
	}
}

/* Writes the DATA LIST statement to the SPS-syntax. */
func DataList(o string, f *os.File, d *Variables) error {
	_, err := f.WriteString(fmt.Sprintf("FILE HANDLE longdata\n/NAME=\"%s\".\n", o))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprint("DATA LIST FILE=longdata\n/"))
	if err != nil {
		return err
	}
	for _, v := range d.Variable {
		if v.Type != "multiple" {
			_, err = f.WriteString(fmt.Sprintf("\t%s\t%d-%d%v\n",
				v.Name, v.Position.Start, v.Position.Finish, v.VarType()))
			if err != nil {
				return err
			}
		} else {
			for i, mult := range v.Vals {
				_, err = f.WriteString(fmt.Sprintf("\t%s#%d\t%d-%d\n",
					v.Name, mult.Value, v.Position.Start+i, v.Position.Start+i))
				if err != nil {
					return err
				}
			}
		}
	}
	_, err = f.WriteString(fmt.Sprint(".\n\n"))
	if err != nil {
		return err
	}
	return nil
}


/* Writes the VARIABLE LABELS statement to the SPS-syntax. */
func VariableLabels(f *os.File, d *Variables) error {
	_, err := f.WriteString(fmt.Sprint("VARIABLE LABELS\n"))
	if err != nil {
		return err
	}
	for _, v := range d.Variable {
		if v.Type != "multiple" {
			_, err = f.WriteString(fmt.Sprintf("\t%s\t\"%s\"\n", v.Name, v.Label))
			if err != nil {
				return err
			}
		} else {
			for _, mult := range v.Vals {
				_, err = f.WriteString(fmt.Sprintf("\t%s#%d\t\"%s\"\n", v.Name, mult.Value, v.Label))
				if err != nil {
					return err
				}
			}
		}
	}
	_, err = f.WriteString(fmt.Sprint(".\nEXECUTE.\n\n\n"))
	if err != nil {
		return err
	}
	return nil
}


/* Writes the VALUE LABELS statement to the SPS-syntax. */
func ValueLabels(f *os.File, d *Variables) error {
	_, err := f.WriteString(fmt.Sprint("VALUE LABELS\n"))
	if err != nil {
		return err
	}
	for _, v := range d.Variable {
		if v.Type == "single" {
			_, err = f.WriteString(fmt.Sprintf("\t%s\n", v.Name))
			if err != nil {
				return err
			}
			for _, vs := range v.Vals {
				_, err = f.WriteString(fmt.Sprintf("\t\t%d \"%s\"\n", vs.Value, v.Name))
				if err != nil {
					return err
				}
			}
			_, err = f.WriteString(fmt.Sprint("/"))
			if err != nil {
				return err
			}
		} else if v.Type == "multiple" {
			for _, mult := range v.Vals {
				_, err = f.WriteString(fmt.Sprintf("\t%s#%d\n", v.Name, mult.Value))
				if err != nil {
					return err
				}
				_, err = f.WriteString(fmt.Sprintf("\t\t0\"No\"\n\t\t1 \"%s\"\n/", mult.Name))
				if err != nil {
					return err
				}
			}
		} else if v.Type == "logical" {
			_, err = f.WriteString(fmt.Sprintf("\t%s\n", v.Name))
			if err != nil {
				return err
			}
			_, err = f.WriteString(fmt.Sprint("\t\t0\"False\"\n\t\t1 \"True\"\n/"))
			if err != nil {
				return err
			}
		}
	}
	_, err = f.WriteString(fmt.Sprint("EXECUTE.\n\n"))
	if err != nil {
		return err
	}
	return nil
}


/* Creates a line to save the SPSS file as a *.sav */
func SaveToSPSS(p string, fn string, f *os.File) error {
	_, err := f.WriteString(fmt.Sprintf("SAVE OUTFILE='%s/%s.sav'\n/COMPRESSED.", p, fn))
	if err != nil {
		log.Fatalln(err)
	}
	return nil
}


func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Usage: XMLtoSPS <XML:filepath> <ASC:filepath>")
	} // Makes sure we have enough arguments to run the program
	input := os.Args[1]
	xmlFile, err := os.Open(input) // Opens the XML file
	if err != nil {
		log.Fatalln(err)
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)
	data := new(Variables)
	xml.Unmarshal(b, &data) // Unmarshals the XML file

	fn := fmt.Sprint(strings.Trim(path.Base(input), path.Ext(input)))
	file, err := os.Create(fmt.Sprintf("%s/%s.sps", path.Dir(input), fn)) // Creates the SPS file
	if err != nil {
		log.Fatalf("Please use forward slash in file path. As an example C:/Users/...\n%v", err)
	}
	defer file.Close()

	err = DataList(os.Args[2], file, data)
	if err != nil {log.Fatalln(err)}

	err = VariableLabels(file, data)
	if err != nil {log.Fatalln(err)}

	err = ValueLabels(file, data)
	if err != nil {log.Fatalln(err)}

	err = SaveToSPSS(path.Dir(input), fn, file)
	if err != nil {log.Fatalln(err)}
}
