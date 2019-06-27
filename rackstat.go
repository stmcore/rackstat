package rackstat

import (
	"encoding/xml"
	"strings"
	"sync"

	"github.com/get-stream-origin/digestauth"
	"github.com/init-data-origin/initdataorigin"
)

type Racks struct {
	Racks []Rack
}

type Rack struct {
	Name     string
	Machines []Machine
}
type Machine struct {
	Name string
	IP   string
	Stat CurrentMachineStatistics
}

type CurrentMachineStatistics struct {
	ServerUptime    int32 `xml:"ServerUptime"`
	CPUIdle         int32 `xml:"CpuIdle"`
	CPUUser         int32 `xml:"CpuUser"`
	CPUSystem       int32 `xml:"CpuSystem"`
	MemoryFree      int64 `xml:"MemoryFree"`
	MemoryUsed      int64 `xml:"MemoryUsed"`
	HeapFree        int64 `xml:"HeapFree"`
	HeapUsed        int64 `xml:"HeapUsed"`
	DiskFree        int64 `xml:"DiskFree"`
	DiskUsed        int64 `xml:"DiskUsed"`
	ConnectionCount int32 `xml:"ConnectionCount"`
}

func (self *Racks) GetAllRackName() []string {
	//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	var initdata initdataorigin.DataOrigins
	servers := initdata.GetServers()
	keys := make(map[string]bool)
	list := []string{}

	for _, server := range servers {
		if _, value := keys[server.Rack]; !value {
			keys[server.Rack] = true
			list = append(list, server.Rack)
		}
	}

	return list
	//c.JSON(http.StatusOK, gin.H{"RackName": list})
}

func (self *Racks) GetStatMachine(rackName string, machine *[]Machine, wg *sync.WaitGroup) {

	defer func() {
		wg.Done()
	}()

	var initdata initdataorigin.DataOrigins
	servers := initdata.GetServers()

	var i = 0
	var size = 0

	for _, server := range servers {
		if rackName == server.Rack {
			size++
		}
	}

	result := make([]Machine, size)

	for _, server := range servers {
		var digest digestauth.Digest
		if rackName == server.Rack {
			result[i].IP = server.IP
			result[i].Name = server.Hostname

			arrIP := strings.Split(server.IP, ".")
			lastTwoIP := strings.Join(arrIP[len(arrIP)-2:], ".")

			url := "http://" + server.IP + ":8087/v2/machine/monitoring/current"
			data, err := digest.GetInfo(url, "sysadm", "1down2go@"+lastTwoIP, "GET")

			if err != nil {
				//log.Println(err)
			}

			xml.Unmarshal([]byte(data), &result[i].Stat)

			i++
		}
	}

	*machine = result
}

func (self *Racks) FetchAllRackStatus() {

	allracks := self.GetAllRackName()
	racks := make([]Rack, len(allracks))

	var i = 0
	wg := &sync.WaitGroup{}

	for _, rackname := range allracks {
		wg.Add(1)
		racks[i].Name = rackname
		go self.GetStatMachine(rackname, &racks[i].Machines, wg)
		i++
	}
	wg.Wait()
	self.Racks = racks

}
