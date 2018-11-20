// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package edgex

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-ui-server/internal/fulcro"
	"github.com/russolsen/transit"

	"gopkg.in/resty.v1"
)

func getDevices() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "device")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "device")
	result = fulcro.Remove(result, "profile", "deviceResources")
	result = fulcro.Remove(result, "profile", "resources")
	result = fulcro.Remove(result, "profile", "commands")
	result = fulcro.MakeKeyword(result, "id")
	result = fulcro.MakeKeyword(result, "adminState")
	result = fulcro.MakeKeyword(result, "operatingState")
	result = fulcro.MakeKeyword(result, "service", "adminState")
	result = fulcro.MakeKeyword(result, "service", "operatingState")
	result = fulcro.MakeKeyword(result, "profile", "id")
	return result
}

func Devices(params []interface{}, args map[interface{}]interface{}) interface{} {
	return fulcro.Keywordize(getDevices())
}

func getDeviceServices() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "deviceservice")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "device-service")
	result = fulcro.MakeKeyword(result, "id")
	result = fulcro.MakeKeyword(result, "adminState")
	result = fulcro.MakeKeyword(result, "operatingState")
	result = fulcro.MakeKeyword(result, "addressable", "id")
	return result
}

func DeviceServices(params []interface{}, args map[interface{}]interface{}) interface{} {
	return fulcro.Keywordize(getDeviceServices())
}

func ScheduleEvents(params []interface{}, args map[interface{}]interface{}) interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "scheduleevent")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "schedule-event")
	result = fulcro.MakeKeyword(result, "id")
	result = fulcro.MakeKeyword(result, "adminState")
	result = fulcro.MakeKeyword(result, "operatingState")
	result = fulcro.MakeKeyword(result, "addressable", "id")
	result = fulcro.Keywordize(result)
	return result
}

func getAddressables() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "addressable")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "addressable")
	result = fulcro.MakeKeyword(result, "id")
	return result
}

func Addressables(params []interface{}, args map[interface{}]interface{}) interface{} {
	return fulcro.Keywordize(getAddressables())
}

func getProfiles() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "deviceprofile")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "device-profile")
	result = fulcro.MakeKeyword(result, "id")
	return result
}

func doGet(getInfo interface{}) interface{} {
	var data map[string]interface{}
	info := getInfo.(map[string]interface{})
	url := info["url"].(string)
	resp, _ := resty.R().Get(url)
	json.Unmarshal(resp.Body(), &data)
	result := make([][2]string, len(data))
	pos := 0
	for k, v := range data {
		result[pos] = [2]string{k, v.(string)}
		pos++
	}
	return result
}

func applyGets(data interface{}) interface{} {
	commands := data.([]map[string]interface{})
	count := 0
	for i, cmd := range commands {
		haveData := false
		get, haveGet := cmd["get"]
		if haveGet {
			delete(commands[i], "get")
			_, haveResp := get.(map[string]interface{})["responses"]
			if haveResp {
				values := doGet(get)
				commands[i]["value"] = values
				count += len(values.([][2]string))
				haveData = true
			}
		}
		if !haveData {
			dummyValue := make([]interface{}, 1)
			dummyValue[0] = [2]string{cmd["name"].(string), "N/A"}
			commands[i]["value"] = dummyValue
			count++
		}
	}
	result := make([]map[string]interface{}, count)
	pos := 0
	for _, cmd := range commands {
		values := cmd["value"].([][2]string)
		for i, val := range values {
			c := make(map[string]interface{})
			for k, v := range cmd {
				c[k] = v
			}
			result[pos] = c
			result[pos]["value"] = val
			result[pos]["pos"] = i
			result[pos]["size"] = len(values)
			pos++
		}
	}
	return result
}

func Commands(params []interface{}, args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resp, _ := resty.R().Get(getEndpoint(ClientCommand) + "device/" + string(id))
	var data map[string]interface{}
	var result interface{}
	json.Unmarshal(resp.Body(), &data)
	commands := data["commands"].([]interface{})
	result = make([]map[string]interface{}, len(commands))
	for i, cmd := range commands {
		result.([]map[string]interface{})[i] = cmd.(map[string]interface{})
	}
	result = fulcro.AddType(result, ClientCommand)
	result = fulcro.MakeKeyword(result, "id")
	result = applyGets(result)
	return fulcro.Keywordize(result)
}

func getReadingsInTimeRange(name string, from int64, to int64) interface{} {
	const batchSize = 100
	const maxRequests = 100
	result := make([]interface{}, batchSize * maxRequests)
	ids := make(map[string]bool)
	pos := 0
	limit := maxRequests
	toStr := strconv.FormatInt(to, 10)
	batchStr := strconv.FormatInt(batchSize, 10)
	var count int
	for ok := true; ok; ok = (count == batchSize) && (limit > 0) {
		fromStr := strconv.FormatInt(from, 10)
		resp, _ := resty.R().Get(getEndpoint(ClientData) + "reading/" + fromStr + "/" + toStr + "/" + batchStr)
		var data []map[string]interface{}
		json.Unmarshal(resp.Body(), &data)
		readings := fulcro.AddType(data, "reading").([]map[string]interface{})
		count = len(readings)
		from = int64(readings[count-1]["created"].(float64))
		for _, reading := range readings {
			if reading["device"].(string) != name {
				continue
			}
			id := reading["id"].(string)
			if !ids[id] {
				ids[id] = true
				result[pos] = fulcro.MakeKeyword(reading, "id")
				pos++
			}
		}
		limit--
	}
	return result[:pos]
}

func DeviceReadings(params []interface{}, args map[interface{}]interface{}) interface{} {
	name := fulcro.GetString(args, "name")
	from := fulcro.GetInt(args, "from")
	to := fulcro.GetInt(args, "to")
	return fulcro.Keywordize(getReadingsInTimeRange(name, from, to))
}

func Profiles(params []interface{}, args map[interface{}]interface{}) interface{} {
	return fulcro.Keywordize(getProfiles())
}

func ProfileYaml(params []interface{}, args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "deviceprofile/yaml/" + string(id))
	m := make(map[string]interface{})
	m["yaml"] = resp.String()
	result := make([]interface{}, 1)
	result[0] = m
	return fulcro.Keywordize(result)
}

func addDefault(data interface{}, keys ...string) interface{} {
	schedules := data.([]map[string]interface{})
	for _, s := range schedules {
		for _, key := range keys {
			v := s[key]
			if v == nil {
				s[key] = 0
			}
		}
	}
	return schedules
}

func getSchedules() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "schedule")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "schedule")
	result = fulcro.MakeKeyword(result, "id")
	result = addDefault(result, "start", "end")
	return result
}

func getScheduleEvents() interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientMetadata) + "scheduleevent")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	result := fulcro.AddType(data, "schedule-event")
	result = fulcro.MakeKeyword(result, "id")
	result = fulcro.MakeKeyword(result, "addressable", "id")
	return result
}

func ShowSchedules(params []interface{}, args map[interface{}]interface{}) interface{} {
	result := make(map[string]interface{})
	result["content"] = getSchedules()
	result["events"] = getScheduleEvents()
	return fulcro.Keywordize(result)
}

func ShowExports(params []interface{}, args map[interface{}]interface{}) interface{} {
	resp, _ := resty.R().Get(getEndpoint(ClientExport) + "registration")
	var data []map[string]interface{}
	json.Unmarshal(resp.Body(), &data)
	exports := fulcro.AddType(data, ClientExport)
	exports = fulcro.MakeKeyword(exports, "id")
	exports = fulcro.MakeKeyword(exports, "destination")
	exports = fulcro.MakeKeyword(exports, "format")
	exports = fulcro.MakeKeyword(exports, "compression")
	exports = fulcro.MakeKeyword(exports, "encryption", "encryptionAlgorithm")
	result := make(map[string]interface{})
	result["content"] = exports
	return fulcro.Keywordize(result)
}

func ShowProfiles(params []interface{}, args map[interface{}]interface{}) interface{} {
	result := make(map[string]interface{})
	result["content"] = getProfiles()
	return fulcro.Keywordize(result)
}

func ShowDevices(params []interface{}, args map[interface{}]interface{}) interface{} {
	result := make(map[string]interface{})
	result["content"] = getDevices()
	result["services"] = getDeviceServices()
	result["schedules"] = getSchedules()
	result["addressables"] = getAddressables()
	result["profiles"] = getProfiles()
	return fulcro.Keywordize(result)
}

func ShowAddressables(params []interface{}, args map[interface{}]interface{}) interface{} {
	result := make(map[string]interface{})
	result["content"] = getAddressables()
	return fulcro.Keywordize(result)
}

func getLogsInTimeRange(from int64, to int64) interface{} {
	const batchSize = 100
	const maxRequests = 100
	result := make([]interface{}, batchSize * maxRequests)
	ids := make(map[transit.Keyword]bool)
	pos := 0
	limit := maxRequests
	toStr := strconv.FormatInt(to, 10)
	batchStr := strconv.FormatInt(batchSize, 10)
	var count int
	for ok := true; ok; ok = (count == batchSize) && (limit > 0) {
		fromStr := strconv.FormatInt(from, 10)
		resp, _ := resty.R().Get(getEndpoint(ClientLogging) + "logs/" + fromStr + "/" + toStr + "/" + batchStr)
		var data []map[string]interface{}
		json.Unmarshal(resp.Body(), &data)
		logs := fulcro.AddId(fulcro.AddType(data, "log-entry")).([]map[string]interface{})
		count = len(logs)
		from = int64(logs[count-1]["created"].(float64))
		for _, entry := range logs {
			id := entry["id"].(transit.Keyword)
			if !ids[id] {
				ids[id] = true
				result[pos] = entry
				pos++
			}
		}
		limit--
	}
	return result[:pos]
}

func ShowLogs(params []interface{}, args map[interface{}]interface{}) interface{} {
	start := fulcro.GetInt(args, "start")
	end := fulcro.GetInt(args, "end")
	result := make(map[string]interface{})
	result["content"] = getLogsInTimeRange(start, end)
	return fulcro.Keywordize(result)
}

func ReadingPage(params []interface{}, args map[interface{}]interface{}) interface{} {
	result := make(map[string]interface{})
	result["devices"] = getDevices()
	return fulcro.Keywordize(result)
}

func UpdateLockMode(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	mode := fulcro.GetKeyword(args, "mode")
	resty.R().Put(getEndpoint(ClientCommand) + "device/" + string(id) + "/adminstate/" + string(mode))
	return id
}

func UploadProfile(args map[interface{}]interface{}) interface{} {
	fileId := fulcro.GetInt(args, "file-id")
	fileName := "tmp-" + strconv.FormatInt(fileId, 10)
	resty.R().
		SetHeader("Content-Type", "application/x-yaml").
		SetFile("file", fileName).
		Post(getEndpoint(ClientMetadata) + "deviceprofile/uploadfile")
	os.Remove(fileName)
	return fileId
}

func DeleteProfile(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientMetadata) + "deviceprofile/id/" + string(id))
	return id
}

type Named struct {
	Name string `json:"name"`
}

type Device struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Labels         []string `json:"labels"`
	Profile        Named    `json:"profile"`
	Service        Named    `json:"service"`
	Addressable    Named    `json:"addressable"`
	AdminState     string   `json:"adminState"`
	OperatingState string   `json:"operatingState"`
}

func AddDevice(args map[interface{}]interface{}) interface{} {
	name := fulcro.GetString(args, "name")
	description := fulcro.GetString(args, "description")
	labels := fulcro.GetStringSeq(args, "labels")
	profileName := fulcro.GetString(args, "profile-name")
	serviceName := fulcro.GetString(args, "service-name")
	addressableName := fulcro.GetString(args, "addressable-name")
	device := Device{Name: name,
		Description:    description,
		Labels:         labels,
		Profile:        Named{Name: profileName},
		Service:        Named{Name: serviceName},
		Addressable:    Named{Name: addressableName},
		AdminState:     "UNLOCKED",
		OperatingState: "ENABLED",
	}
	resty.R().SetBody(device).Post(getEndpoint(ClientMetadata) + "device")
	return nil
}

func DeleteDevice(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientMetadata) + "device/id/" + string(id))
	return id
}

type Addressable struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address"`
	Protocol  string `json:"protocol"`
	Port      int64  `json:"port"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Publisher string `json:"publisher"`
	Topic     string `json:"topic"`
	User      string `json:"user"`
	Password  string `json:"password"`
}

func AddAddressable(args map[interface{}]interface{}) interface{} {
	tempid := fulcro.GetTempId(args, "tempid")
	name := fulcro.GetString(args, "name")
	address := fulcro.GetString(args, "address")
	protocol := fulcro.GetString(args, "protocol")
	port := fulcro.GetInt(args, "port")
	path := fulcro.GetString(args, "path")
	method := strings.ToUpper(string(fulcro.GetKeyword(args, "method")))
	publisher := fulcro.GetString(args, "publisher")
	topic := fulcro.GetString(args, "topic")
	user := fulcro.GetString(args, "user")
	password := fulcro.GetString(args, "password")
	addressable := Addressable{
		Id:        "",
		Name:      name,
		Address:   address,
		Protocol:  protocol,
		Port:      port,
		Path:      path,
		Method:    method,
		Publisher: publisher,
		Topic:     topic,
		User:      user,
		Password:  password,
	}
	resp, _ := resty.R().SetBody(addressable).Post(getEndpoint(ClientMetadata) + "addressable")
	return fulcro.MkTempIdResult(tempid, resp)
}

func EditAddressable(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	address := fulcro.GetString(args, "address")
	protocol := fulcro.GetString(args, "protocol")
	port := fulcro.GetInt(args, "port")
	path := fulcro.GetString(args, "path")
	method := strings.ToUpper(string(fulcro.GetKeyword(args, "method")))
	publisher := fulcro.GetString(args, "publisher")
	topic := fulcro.GetString(args, "topic")
	user := fulcro.GetString(args, "user")
	password := fulcro.GetString(args, "password")
	addressable := Addressable{
		Id:        string(id),
		Name:      "",
		Address:   address,
		Protocol:  protocol,
		Port:      port,
		Path:      path,
		Method:    method,
		Publisher: publisher,
		Topic:     topic,
		User:      user,
		Password:  password,
	}
	resty.R().SetBody(addressable).Put(getEndpoint(ClientMetadata) + "addressable")
	return id
}

func DeleteAddressable(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientMetadata) + "addressable/id/" + string(id))
	return id
}

type Schedule struct {
	Name      string `json:"name,omitempty"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Frequency string `json:"frequency"`
	RunOnce   bool   `json:"run-once"`
}

func AddSchedule(args map[interface{}]interface{}) interface{} {
	tempid := fulcro.GetTempId(args, "tempid")
	name := fulcro.GetString(args, "name")
	start := fulcro.GetInt(args, "start")
	end := fulcro.GetInt(args, "end")
	frequency := fulcro.GetString(args, "frequency")
	runOnce := fulcro.GetBool(args, "run-once")
	schedule := Schedule{
		Name:      name,
		Start:     start,
		End:       end,
		Frequency: frequency,
		RunOnce:   runOnce,
	}
	resp, _ := resty.R().SetBody(schedule).Post(getEndpoint(ClientMetadata) + "schedule")
	return fulcro.MkTempIdResult(tempid, resp)
}

func DeleteSchedule(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientMetadata) + "schedule/id/" + string(id))
	return id
}

type ScheduleEvent struct {
	Name        string `json:"name,omitempty"`
	Addressable Named  `json:"addressable"`
	Parameters  string `json:"parameters"`
	Schedule    string `json:"schedule"`
	Service     string `json:"service"`
}

func AddScheduleEvent(args map[interface{}]interface{}) interface{} {
	tempid := fulcro.GetTempId(args, "tempid")
	name := fulcro.GetString(args, "name")
	addressableName := fulcro.GetString(args, "addressable-name")
	parameters := fulcro.GetString(args, "parameters")
	schedule := fulcro.GetString(args, "schedule-name")
	service := fulcro.GetString(args, "service-name")
	scheduleEvent := ScheduleEvent{
		Name:        name,
		Addressable: Named{Name: addressableName},
		Parameters:  parameters,
		Schedule:    schedule,
		Service:     service,
	}
	resp, _ := resty.R().SetBody(scheduleEvent).Post(getEndpoint(ClientMetadata) + "scheduleevent")
	return fulcro.MkTempIdResult(tempid, resp)
}

func DeleteScheduleEvent(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientMetadata) + "scheduleevent/id/" + string(id))
	return id
}

type Encryption struct {
	EncryptionAlgorithm string `json:encryptionAlgorithm"`
	EncryptionKey       string `json:"encryptionKey,omitempty"`
	InitializingVector  string `json:"initializingVector,omitempty"`
}

type Export struct {
	Id          string `json:"id,omitempty"`
	Addr        Addressable
	Format      string `json:"format"`
	Destination string `json:"destination"`
	Compression string `json:"compression"`
	Encrypt     Encryption
	Enable      bool `json:"enable"`
}

func AddExport(args map[interface{}]interface{}) interface{} {
	tempid := fulcro.GetTempId(args, "tempid")
	addressable := fulcro.GetMap(args, "addressable")
	export := Export{
		Id: "",
		Addr: Addressable{
			Id:        "",
			Name:      fulcro.GetString(addressable, "name"),
			Address:   fulcro.GetString(addressable, "address"),
			Protocol:  fulcro.GetString(addressable, "protocol"),
			Port:      fulcro.GetInt(addressable, "port"),
			Path:      fulcro.GetString(addressable, "path"),
			Method:    fulcro.GetString(addressable, "method"),
			Publisher: fulcro.GetString(addressable, "publisher"),
			Topic:     fulcro.GetString(addressable, "topic"),
			User:      fulcro.GetString(addressable, "user"),
			Password:  fulcro.GetString(addressable, "password"),
		},
		Format:      fulcro.GetKeywordAsString(args, "format"),
		Destination: fulcro.GetKeywordAsString(args, "destination"),
		Compression: fulcro.GetKeywordAsString(args, "compression"),
		Encrypt: Encryption{
			EncryptionAlgorithm: fulcro.GetKeywordAsString(args, "encryptionAlgorithm"),
			EncryptionKey:       fulcro.GetKeywordAsString(args, "encryptionKey"),
			InitializingVector:  fulcro.GetKeywordAsString(args, "initializingVector"),
		},
		Enable: fulcro.GetBool(args, "enable"),
	}
	resp, _ := resty.R().SetBody(export).Post(getEndpoint(ClientExport) + "registration")
	return fulcro.MkTempIdResult(tempid, resp)
}

func EditExport(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	addressable := fulcro.GetMap(args, "addressable")
	export := Export{
		Id: string(id),
		Addr: Addressable{
			Id:        "",
			Name:      fulcro.GetString(addressable, "name"),
			Address:   fulcro.GetString(addressable, "address"),
			Protocol:  fulcro.GetString(addressable, "protocol"),
			Port:      fulcro.GetInt(addressable, "port"),
			Path:      fulcro.GetString(addressable, "path"),
			Method:    fulcro.GetString(addressable, "method"),
			Publisher: fulcro.GetString(addressable, "publisher"),
			Topic:     fulcro.GetString(addressable, "topic"),
			User:      fulcro.GetString(addressable, "user"),
			Password:  fulcro.GetString(addressable, "password"),
		},
		Format:      fulcro.GetKeywordAsString(args, "format"),
		Destination: fulcro.GetKeywordAsString(args, "destination"),
		Compression: fulcro.GetKeywordAsString(args, "compression"),
		Encrypt: Encryption{
			EncryptionAlgorithm: fulcro.GetKeywordAsString(args, "encryptionAlgorithm"),
			EncryptionKey:       fulcro.GetKeywordAsString(args, "encryptionKey"),
			InitializingVector:  fulcro.GetKeywordAsString(args, "initializingVector"),
		},
		Enable: fulcro.GetBool(args, "enable"),
	}
	resty.R().SetBody(export).Put(getEndpoint(ClientExport) + "registration")
	return id
}

func DeleteExport(args map[interface{}]interface{}) interface{} {
	id := fulcro.GetKeyword(args, "id")
	resty.R().Delete(getEndpoint(ClientExport) + "registration/id/" + string(id))
	return id
}

func getValueSeq(args map[interface{}]interface{}, id string) [][]interface{} {
	outer := args[transit.Keyword(id)].([]interface{})
	result := make([][]interface{}, len(outer))
	for i, s := range outer {
		seq := s.([]interface{})
		result[i] = make([]interface{}, len(seq))
		for j, v := range seq {
			result[i][j] = v
		}
	}
	return result
}

func IssueSetCommand(args map[interface{}]interface{}) interface{} {
	url := fulcro.GetString(args, "url")
	values := getValueSeq(args, "values")
	data := make(map[string]interface{}, len(values))
	for _, v := range values {
		data[v[0].(string)] = v[2]
	}
	resty.R().SetBody(data).Put(url)
	return nil
}
