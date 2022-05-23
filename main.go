package main

// API Key for Meraki Sandbox 6bec40cf957de430a6f1f2baa056b99a4fac9ea0
// default baseUrl for Meraki API https://api.meraki.com/api/v1
import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	// "io"
	// "strconv"
)

// Create the JSON Structure - https://pkg.go.dev/encoding/json for documentation

type SerialPort struct {
	Serial  string `json:"serial"`
	PortNum string `json:"portnum"`
}
type ConfigPayload struct {
	Name string `json:"name"`
	//Tags                    *Tags  `json:"tags,omitempty"`
	//Tags                    string `json:"tags"`
	Enabled                 string `json:"enabled"`
	Porttype                string `json:"type"`
	Vlan                    string `json:"vlan"`
	VoiceVlan               string `json:"voiceVlan,omitempty"`
	AllowedVlans            string `json:"allowedVlans,omitempty"`
	PoeEnabled              string `json:"poeEnabled,omitempty"`
	IsolationEnabled        string `json:"isolationEnabled,omitempty"`
	RstpEnabled             string `json:"rstpEnabled,omitempty"`
	StpGuard                string `json:"stpGuard,omitempty"`
	LinkNegotiation         string `json:"linkNegotiation,omitempty"`
	PortScheduleId          string `json:"portScheduleID,omitempty"`
	UdId                    string `json:"udId,omitempty"`
	AccessPolicyType        string `json:"accessPolicyType,omitempty"`
	AccessPolicyNumber      string `json:"accessPolicyNumber,omitempty"`
	MacAllowList            string `json:"macAllowList,omitempty"`
	StickyMacAllowList      string `json:"stickyMacAllowList,omitempty"`
	StickyMacAllowListLimit string `json:"stickyMacAllowListLimit,omitempty"`
	StormControlEnabled     string `json:"stormControlEnabled,omitempty"`
}

// type Tags struct {
// 	tag string `json:"tags"`
// }

var defaultfile string = "./MerakiSwitchPortCSV.csv"
var file string
var defaulturl string = "https://api.meraki.com/api/v1"
var apiurl string

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}

func getcsvfile() {

	fileinput := bufio.NewScanner(os.Stdin)

	fmt.Println("\n\nType the full path to the CSV file you would like to use")
	fmt.Println("Press Enter to use default MerakiSwitchPortCSV.csv in current directory")
	fmt.Println("NOTE: The first line of the CSV file will be ignored. It should be a header row.")
	fmt.Println("--------------------------------------------------------------------------------------")
	fmt.Print("-> ")

	for fileinput.Scan() {

		fileinput := fileinput.Text()

		file = fileinput

		if strings.Compare("", file) == 0 {
			file = defaultfile
		}
		if !fileExists(file) {
			fmt.Println("\n", file, " - File does not exist or is a directory")
			fmt.Print("-> ")
			continue
		} else {
			break
		}
	}
	return
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// func readSample(rs io.ReadSeeker) ([][]string, error) {
// 	// Skip first row (line)
// 	row1, err := bufio.NewReader(rs).ReadSlice('\n')
// 	if isError(err) {
// 		return nil, err
// 	}
// 	_, err = rs.Seek(int64(len(row1)), io.SeekStart)
// 	if isError(err) {
// 		return nil, err
// 	}

// 	// Read remaining rows
// 	r := csv.NewReader(rs)
// 	//r.Comma = ';'
// 	rows, err := r.ReadAll()
// 	if isError(err) {
// 		return nil, err
// 	}
// 	return rows, nil
// }

func main() {

	// client trace to log whether the request's underlying tcp connection was re-used
	// clientTrace := &httptrace.ClientTrace{
	// 	GotConn: func(info httptrace.GotConnInfo) { log.Printf("conn was reused: %t", info.Reused) },
	// }
	// traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	debugPtr := flag.Bool("debug", false, "enable debug to only print what would be sent to the API")

	flag.Parse()

	//var body []byte
	var response *http.Response
	var request *http.Request

	getcsvfile()

	apireader := bufio.NewScanner(os.Stdin)
	fmt.Println("\nWhat is the base URL for the Meraki API?")
	fmt.Println("Press Enter to accept the default of https://api.meraki.com/api/v1")
	fmt.Print("-> ")
	apireader.Scan()
	apiurl := apireader.Text()

	if strings.Compare("", apiurl) == 0 {
		apiurl = defaulturl
	}

	keyreader := bufio.NewScanner(os.Stdin)
	fmt.Println("\nWhat is the API Key? ")
	fmt.Println("(example Meraki API key 6bec40cf957de430a6f1f2baa056b99a4fac9ea0)")
	fmt.Print("-> ")
	keyreader.Scan()
	apikey := keyreader.Text()

	// fmt.Println("Base URL is: ", apiurl)
	// fmt.Println("API Key is: ", apikey)

	csv_file, err := os.Open(file) //the _noheader.csv doesn't work - gets a panic runtime error because there are indexs out of range
	// this was because the no headers file didn't have enough commas - hence the index out of range error. source has been updated and we can discuss the long term desire / implications
	// I'm sure I can use the header file, which will auto create all the necessary commas and in the api loop drop the [0] record from the loop
	if isError(err) {
		return
	}

	defer csv_file.Close()

	//records, err := readSample(csv_file)
	r := csv.NewReader(csv_file)
	r.Comma = ','
	r.Comment = '#'
	records, err := r.ReadAll()
	if isError(err) {
		return
	}

	// Check CSV file for duplicate Serial Number + Port Number combo's
	nameExistMap := make(map[string]bool)

	for _, row := range records {
		name := row[0] + row[1]

		if _, exist := nameExistMap[name]; exist {
			fmt.Println("Serial Number:", row[0], "and Port Number:", row[1], "appears to be a duplicate in the CSV data")
			os.Exit(1)
		} else {
			nameExistMap[name] = true
			continue
		}
	}

	var switchportdata SerialPort
	var switchportdatas []SerialPort
	var configpayload ConfigPayload
	var configpayloads []ConfigPayload
	//var tagarray TagArray

	for _, rec := range records {
		switchportdata.Serial = rec[0]
		switchportdata.PortNum = rec[1]
		configpayload.Name = rec[2]
		// All of this mess was Aaron trying to figure out how to send tags as an array. For now we have
		// decided to not send Tags at all.
		// tagarray := strings.Split(rec[3], ";")
		// fmt.Println(tagarray)
		//for _, tag := range tagarray {
		//	configpayload.Tags = tagarray.tag
		//}
		//tagarray.tag = []TagArray{strings.Split(rec[3], ";")}

		//	configpayload.Tags = &Tags{
		//		tag: rec[3],
		//	}
		//}
		//configpayload.Tags = rec[3]
		configpayload.Enabled = rec[4]
		configpayload.Porttype = rec[5]
		configpayload.Vlan = rec[6]
		configpayload.VoiceVlan = rec[7]
		configpayload.AllowedVlans = rec[8]
		configpayload.PoeEnabled = rec[9]
		configpayload.IsolationEnabled = rec[10]
		configpayload.RstpEnabled = rec[11]
		configpayload.StpGuard = rec[12]
		configpayload.LinkNegotiation = rec[13]
		configpayload.PortScheduleId = rec[14]
		configpayload.UdId = rec[15]
		configpayload.AccessPolicyType = rec[16]
		configpayload.AccessPolicyNumber = rec[17]
		configpayload.MacAllowList = rec[18]
		configpayload.StickyMacAllowList = rec[19]
		configpayload.StickyMacAllowListLimit = rec[20]
		configpayload.StormControlEnabled = rec[21]
		configpayloads = append(configpayloads, configpayload)
		switchportdatas = append(switchportdatas, switchportdata)
		//marshal and print json data for each record

		// fmt.Print(apiurl)
		// fmt.Print("URI TARGET:  https://api.meraki.com/api/v1", "/devices/", switchportdata.Serial, "/switch/ports/", switchportdata.PortNum, "\n")

		cfpl_json, err := json.Marshal(configpayload)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		//fmt.Println("JSON BODY FOR POST:", string(cfpl_json))

		fullapi := fmt.Sprintf("%s%s%s%s%s", apiurl, "/devices/", switchportdata.Serial, "/switch/ports/", switchportdata.PortNum)
		// fmt.Println(fullapi)
		//client := &http.Client{}

		// print json data
		fmt.Print("\n\n")
		fmt.Println("Sending the below JSON Port Configuration to SN:", switchportdata.Serial, "for Port:", switchportdata.PortNum)
		fmt.Println("API URL:", fullapi)
		fmt.Println(string(cfpl_json))

		// Use this method to trace the HTTP call and determine if the connection is being reused
		// request, err = http.NewRequestWithContext(traceCtx, http.MethodPut, fullapi, bytes.NewBuffer(cfpl_json))
		request, err = http.NewRequest(http.MethodPut, fullapi, bytes.NewBuffer(cfpl_json))
		if err != nil {
			log.Fatalf("HTTP call failed: %s", err)
		}

		if *debugPtr == false {
			request.Header.Add("X-Cisco-Meraki-API-Key", apikey)
			request.Header.Add("Content-Type", "application/json")
			response, err = (&http.Client{}).Do(request)
			if response.StatusCode != http.StatusOK {
				fmt.Println("ERROR - Non-OK HTTP Status:", response.StatusCode)
				body, _ := ioutil.ReadAll(response.Body)
				fmt.Println(string(body))
			} else {
				fmt.Println("SUCCESS")
			}
			if _, err := io.Copy(ioutil.Discard, response.Body); err != nil {
				log.Fatal(err)
			}
			response.Body.Close()
		}

		// Only run below if debug is false
		// if *debugPtr == false {

		// 	// if err != nil {
		// 	// 	log.Fatalf("HTTP call failed: %s", err)
		// 	// }
		// 	defer response.Body.Close()
		// 	// if response.StatusCode != http.StatusOK {
		// 	// 	fmt.Println("ERROR - Non-OK HTTP Status:", response.StatusCode)
		// 	// 	body, _ := ioutil.ReadAll(response.Body)
		// 	// 	fmt.Println(string(body))
		// 	// } else {
		// 	// 	fmt.Println("SUCCESS")
		// 	// }
		// }
		// if err != nil {
		// 	print(err)
		// }
		// fmt.Print(client)
		// fmt.Print(resp, "\n", "\n", "\n")
		// fmt.Print(resp.Body)
		// res, err := client.Do(resp)
		// if err != nil {
		// 	// handle error
		// 	log.Fatal(err)
		// }
		//body, err := ioutil.ReadAll(response.Body)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		//fmt.Println(string(body))
		// defer resp.Body.Close()

		// request := gorequest.New()
		// resp, body, errs := request.Post(apiurl).
		//   Set("X-Cisco-Meraki-API-Key", apikey).
		//   Send(string(cfpl_json)).
		//   End()
		//   if errs != nil {
		//        print(err)
		//      }

		// fmt.Print(body)
		// fmt.Print(resp, "\n", "\n")
		// fmt.Print(request)

	}

	json_data, err := json.Marshal(configpayloads)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// temp_json_data := json.Unmarshal([]byte(json_data), &records)
	// fmt.Print(temp_json_data)

	// var test []Switch
	// err = json.Unmarshal([]byte(json_data), &test)
	// if err != nil {
	//   log.Fatal(err)
	// }
	// fmt.Println(test [15])
	// if errd != nil {
	//   log.Fatal(err)

	//   for _, json_data := range records {
	//       fmt.Print (data, "\n")
	//   }

	// print json data
	// fmt.Println(string(json_data))
	json_file, err := os.Create("sample.json")
	if err != nil {
		fmt.Println(err)
	}
	defer json_file.Close()

	json_file.Write(json_data)
	json_file.Close()

}
