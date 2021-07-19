package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	// flags
	incpatch bool
	incminor bool
	incmajor bool

	file       string
	enviroment string

	property string

	versionChanged bool

	// version number
	prerelease string

	// the main file to store the version to
	yamlFile string
	version  versionConfig
)

type versionConfig struct {
	Major      int    `yaml:"major"`
	Minor      int    `yaml:"minor"`
	Patch      int    `yaml:"patch"`
	Prerelease string `yaml:"prerelease"`
}

func (v *versionConfig) SemanticString() string {
	versionString := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		versionString = fmt.Sprintf("%s-%s", versionString, v.Prerelease)
	}
	return versionString
}

func (v *versionConfig) SemanticStringWOPrerelease() string {
	versionString := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	return versionString
}

func init() {
	// variables for parameter override
	flag.BoolVarP(&incpatch, "inc", "i", false, "increment patch version.")
	flag.BoolVarP(&incminor, "incminor", "m", false, "increment minor version.")
	flag.BoolVarP(&incmajor, "incmajor", "n", false, "increment major version.")

	flag.StringVarP(&prerelease, "pre", "p", "", "prerelease string")

	flag.StringVarP(&file, "file", "f", "", "the file to set the verion into")
	flag.StringVarP(&enviroment, "env", "e", "", "the enviroment to set the version to. \r\nnpm: using npm package.json file\r\nvs: visual studio file\r\niss: inno setup version\r\ngores: winres integration for go")
	flag.StringVarP(&property, "prop", "o", "", "the property for the dfferent env.. \r\niss: inno setup define variable for the version\r\ngores: list of all key pathes")
}

func main() {
	versionChanged = false

	loadYamlFile()

	flag.Parse()

	if incmajor {
		version.Major += 1
		fmt.Printf("increment major: %d\r\n", version.Major)
		versionChanged = true
	}

	if incminor {
		version.Minor += 1
		fmt.Printf("increment minor: %d\r\n", version.Minor)
		versionChanged = true
	}

	if incpatch {
		version.Patch += 1
		fmt.Printf("increment patch: %d\r\n", version.Patch)
		versionChanged = true
	}

	if prerelease != "" {
		if prerelease == " " {
			version.Prerelease = ""
			fmt.Println("prerelease resetted")
			versionChanged = true
		} else {
			version.Prerelease = prerelease
			fmt.Printf("prerelease changed: %s\r\n", version.Prerelease)
			versionChanged = true
		}
	}

	if versionChanged {
		saveYamlFile()
	}

	fmt.Printf("version for processing: %s\r\n", version.SemanticString())
	processVersion()
}

func loadYamlFile() {
	yamlFile = "version.yaml"
	_, err := os.Stat(yamlFile)
	if os.IsNotExist(err) {
		version = versionConfig{
			Major:      0,
			Minor:      0,
			Patch:      1,
			Prerelease: "",
		}
		saveYamlFile()
	}
	data, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Printf("error reading version file. %s\r\n", yamlFile)
		os.Exit(0)
	}
	err = yaml.Unmarshal(data, &version)
	if err != nil {
		fmt.Printf("can't unmarshal config file: %s\r\n", err.Error())
	}

	fmt.Printf("actual Version: %s\r\n", version.SemanticString())
}

func saveYamlFile() {
	f, err := os.Create(yamlFile)
	if err != nil {
		fmt.Printf("error marshaling version. %v\r\n", err)
		return
	}
	defer f.Close()
	err = yaml.NewEncoder(f).Encode(version)
	if err != nil {
		fmt.Printf("error marshaling version. %v\r\n", err)
		return
	}
}

func processVersion() {
	if enviroment != "" {
		switch enviroment {
		case "iss":
			SetInnoSetupVersion()
		case "npm":
			SetNpmVersion()
		case "vs":
			SetVisualStudioVersion()
		case "gores":
			SetGoResVersion()
		case "go":
			SetGoVersion()
		case "ino":
			SetInoVersion()
		case "txt":
			SetTxtVersion()
		}
	}
}

func SetTxtVersion() {
	fmt.Println("changing txt version")

	if property == "" {
		fmt.Println("no property found")
		return
	}

	parts := strings.Split(property, ",")
	line := parts[1]
	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(property)

	line = fmt.Sprintf(line, version.SemanticStringWOPrerelease())

	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	sliceData := strings.Split(string(data), "\n")

	sliceData[lineNumber] = line
	f, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	for x, dataLine := range sliceData {
		_, err := f.WriteString(dataLine)
		if err != nil {
			fmt.Printf("failed writing txt file: %v", err)
			return
		}
		if x < (len(sliceData) - 1) {
			f.WriteString("\n")
		}
	}
}

func SetInoVersion() {
	fmt.Println("changing ino version")

	if property == "" {
		fmt.Println("no property found")
		return
	}

	//parts := strings.Split(property, ",")

	property = strings.ReplaceAll(property, "\\r", "\r")
	property = strings.ReplaceAll(property, "\\n", "\n")
	fmt.Println(property)

	jsonStr := fmt.Sprintf(property, version.SemanticStringWOPrerelease())

	err := ioutil.WriteFile(file, []byte(jsonStr), 0666)
	if err != nil {
		fmt.Printf("failed writing json: %v", err)
		return
	}
}

func SetGoVersion() {
	fmt.Println("changing go version")

	jsonStr := fmt.Sprintf("{\"Major\":%d,\"Minor\":%d,\"Patch\":%d,\"Special\":\"%s\"}", version.Major, version.Minor, version.Patch, version.Prerelease)

	err := ioutil.WriteFile(file, []byte(jsonStr), 0666)
	if err != nil {
		fmt.Printf("failed writing json: %v", err)
		return
	}
}

func SetGoResVersion() {
	fmt.Println("changing go-winres version")

	if property == "" {
		fmt.Println("no property found")
		return
	}

	parts := strings.Split(property, ",")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("can't find gores file: %s", file)
		return
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("failed opening file: %v", err)
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Printf("failed unmarshall json: %v", err)
		return
	}
	var start map[string]interface{}
	for _, property := range parts {
		subParts := strings.Split(property, "/")
		start = jsonData
		for x, key := range subParts {
			if x < len(subParts)-1 {
				ok := false
				start, ok = start[key].(map[string]interface{})
				if !ok {
					fmt.Printf("can't find key %s for property %s", key, property)
					break
				}
			} else {
				_, ok := start[key]
				if !ok {
					fmt.Printf("can't find key %s for property %s", key, property)
					break
				} else {
					start[key] = version.SemanticString()
				}
			}
		}
		fmt.Println()
	}
	jsonStr, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Printf("failed marshall json: %v", err)
		return
	}

	err = ioutil.WriteFile(file, jsonStr, 0666)
	if err != nil {
		fmt.Printf("failed writing json: %v", err)
		return
	}
}

func SetVisualStudioVersion() {
	fmt.Println("changing visual studio assmbly version")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("can't find vs file: %s", file)
		return
	}
	vsfile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	inPropertyGroup := false
	var buf bytes.Buffer
	decoder := xml.NewDecoder(vsfile)
	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error getting token: %v\n", err)
			break
		}

		switch v := token.(type) {
		case xml.StartElement:
			//var myVersion xmlVersion
			if v.Name.Local == "PropertyGroup" {
				inPropertyGroup = true
			}
			if v.Name.Local == "Version" {
				if inPropertyGroup {
					var value string
					if err = decoder.DecodeElement(&value, &v); err != nil {
						fmt.Printf("error decoding xml: %v\r\n", err)
						return
					}
					// modify the version value and encode the element back
					value = version.SemanticStringWOPrerelease()
					if err = encoder.EncodeElement(value, v); err != nil {
						fmt.Printf("error decoding xml: %v\r\n", err)
						return
					}
					continue
				}
			}
		case xml.EndElement:
			if v.Name.Local == "PropertyGroup" {
				inPropertyGroup = false
			}
		}

		if err := encoder.EncodeToken(xml.CopyToken(token)); err != nil {
			log.Fatal(err)
		}
	}

	vsfile.Close()

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(file, buf.Bytes(), 0x666)
	if err != nil {
		fmt.Printf("failed writing xml: %v", err)
		return
	}
}

func SetNpmVersion() {
	fmt.Println("changing npm version")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("can't find npm file: %s", file)
		return
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("failed opening file: %v", err)
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Printf("failed unmarshall json: %v", err)
		return
	}

	_, ok := jsonData["version"]
	if !ok {
		fmt.Println("version field not present")
		return
	}
	jsonData["version"] = version.SemanticString()

	jsonStr, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Printf("failed marshall json: %v", err)
		return
	}

	err = ioutil.WriteFile(file, jsonStr, 0666)
	if err != nil {
		fmt.Printf("failed writing json: %v", err)
		return
	}
}

func SetInnoSetupVersion() {
	fmt.Println("changing inno setup version")
	if property == "" {
		fmt.Println("property not set.")
		return
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("can't find inno setup file: %s", file)
		return
	}
	issfile, err := os.Open(file)
	if err != nil {
		fmt.Printf("failed opening file: %v", err)
		return
	}

	scanner := bufio.NewScanner(issfile)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == ";version number set by GoVersionSetter." {
			continue
		}
		if strings.HasPrefix(line, "#define") {
			parts := strings.Split(line, " ")
			if strings.EqualFold(parts[1], property) {
				found = true
				txtlines = append(txtlines, ";version number set by GoVersionSetter.")
				line = fmt.Sprintf("#define %s \"%s\"", property, version.SemanticString())
			}
		}
		txtlines = append(txtlines, line)
	}

	issfile.Close()

	if !found {
		line := fmt.Sprintf("#define %s \"%s\"", property, version.SemanticString())
		txtlines = append([]string{";version number set by GoVersionSetter.", line}, txtlines...)
	}

	issfile, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer issfile.Close()

	datawriter := bufio.NewWriter(issfile)

	for _, data := range txtlines {
		_, _ = datawriter.WriteString(data + "\r\n")
	}

	datawriter.Flush()
}
