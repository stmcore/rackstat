package rackstat

import (
	"encoding/xml"
	"strings"
	"sync"

	"github.com/stmcore/digestauth"
	"github.com/stmcore/initdataorigin"
)

//Sites data
type Sites struct {
	Sites []Site
}

//Site data
type Site struct {
	Name  string
	Racks []Rack
}

//Racks data
// type Racks struct {
// 	Racks []Rack
// }

//Rack data
type Rack struct {
	Name     string
	Machines []Machine
}

//Machine data
type Machine struct {
	Name string
	IP   string
	Stat CurrentMachineStatistics
}

//CurrentMachineStatistics data
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

//GetAllSiteName uniq site name
func (sites *Sites) GetAllSiteName() []string {
	//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	var initdata initdataorigin.DataOrigins
	servers := initdata.GetServers()
	keys := make(map[string]bool)
	list := []string{}

	for _, server := range servers {
		if _, value := keys[server.Site]; !value {
			keys[server.Site] = true
			list = append(list, server.Site)
		}
	}

	return list
	//c.JSON(http.StatusOK, gin.H{"RackName": list})
}

//GetAllRackName uniq rack name
// func (sites *Sites) GetAllRackName() []string {
// 	//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 	var initdata initdataorigin.DataOrigins
// 	servers := initdata.GetServers()
// 	keys := make(map[string]bool)
// 	list := []string{}

// 	for _, server := range servers {
// 		if _, value := keys[server.Rack]; !value {
// 			keys[server.Rack] = true
// 			list = append(list, server.Rack)
// 		}
// 	}

// 	return list
// 	//c.JSON(http.StatusOK, gin.H{"RackName": list})
// }

//GetRackNameBySite get rack name by site name
func (sites *Sites) GetRackNameBySite(siteName string) []string {
	//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	var initdata initdataorigin.DataOrigins
	servers := initdata.GetServers()
	keys := make(map[string]bool)
	list := []string{}

	for _, server := range servers {
		if server.Site == siteName {
			if _, value := keys[server.Rack]; !value {
				keys[server.Rack] = true
				list = append(list, server.Rack)
			}
		}
	}

	return list
	//c.JSON(http.StatusOK, gin.H{"RackName": list})
}

//GetStatMachine get stat machine from origin api
func (sites *Sites) GetStatMachine(rackName string, machine *[]Machine, wg *sync.WaitGroup) {

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
			data, err := digest.GetInfo(url, "sysadm", "C0rE#"+lastTwoIP, "GET")

			if err != nil {
				//log.Println(err)
			}

			xml.Unmarshal([]byte(data), &result[i].Stat)

			i++
		}
	}

	*machine = result
}

//FetchAllRackStatus update rack status
func (sites *Sites) FetchAllRackStatus() {

	allsites := sites.GetAllSiteName()
	//allracks := sites.GetAllRackName()
	siteArr := make([]Site, len(allsites))

	var i = 0
	wg := &sync.WaitGroup{}

	for _, sitename := range allsites {
		var j = 0

		siteArr[i].Name = sitename
		racks := sites.GetRackNameBySite(sitename)

		rackArr := make([]Rack, len(racks))

		for _, rackname := range racks {
			wg.Add(1)
			rackArr[j].Name = rackname
			go sites.GetStatMachine(rackname, &rackArr[j].Machines, wg)
			j++
		}
		wg.Wait()
		siteArr[i].Racks = rackArr
		i++
	}
	sites.Sites = siteArr
}
