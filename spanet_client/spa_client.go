package spanet_client

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type SpaPumpAttributes struct {
	Id        int
	Active    float64
	Installed string
	Ok        float64
}

type SpaBlowerAttributes struct {
	Mode  float64
	Speed float64
}

type SpaLightsAttributes struct {
	Active     float64
	Mode       float64
	Brightness float64
	Speed      float64
	Colour     float64
}

type SpaFiltrationSettings struct {
	Hour     float64
	Interval float64
}
type SpaSettings struct {
	Mode       string
	Filtration SpaFiltrationSettings
	Lock       float64
}

type SpaAttributes struct {
	TargetTemperature float64
	WaterTemperature  float64
	Heating           float64
	Cleaning          float64
	Sanitising        float64
	Auto              float64
	Sleeping          float64
	Pumps             [5]SpaPumpAttributes
	Blower            SpaBlowerAttributes
	Lights            SpaLightsAttributes
	Settings          SpaSettings
}

type SpaConn struct {
	Socket SpanetSocket
	Conn   net.Conn
}

func (s *SpaConn) setTimeout(timeout float64) {
	if timeout == 0 {
		s.Conn.SetDeadline(time.Time{})
	}
	s.Conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (s *SpaConn) Connect() error {

	var err error
	s.Conn, err = net.DialTimeout("tcp", s.Socket.SpaUrl, time.Duration(5)*time.Second)
	if err != nil {
		return err
	}

	s.setTimeout(5)

	fmt.Println("Opened TCP socket to spa")

	connString := fmt.Sprintf("<connect--%d--%d>", s.Socket.SocketId, s.Socket.MemberId)
	//fmt.Println(connString)

	_, err = s.Conn.Write([]byte(connString))
	if err != nil {
		s.Conn.Close()
		s.Conn = nil
		fmt.Println("Failed to send data to spa")
		return err
	}

	// Check that data sent successfully
	reply := make([]byte, 22)
	s.Conn.Read(reply)
	if string(reply) != "Successfully connected" {
		s.Conn.Close()
		s.Conn = nil
		return errors.New("Failed to handshake with spa" + string(reply))
	}

	//fmt.Println("Successfully connected to spa, ready to send/recieve commands")
	// Reset timeout
	s.setTimeout(0)

	return nil
}

func getStringAttribute(data map[string][]string, row string, col int) string {
	if val, ok := data[row]; ok {
		if len(val) >= col {
			return val[col]
		} else {
			fmt.Println(fmt.Sprintf("Line %s has no col %d, len %d", row, col, len(val)))
		}
	} else {
		fmt.Println("Data has no row: " + row)
	}

	return ""
}

func getFloatAttribute(data map[string][]string, row string, col int) float64 {
	v, _ := strconv.ParseFloat(getStringAttribute(data, row, col), 64)
	return v
}

func getBoolAttribute(data map[string][]string, row string, col int) float64 {
	v, _ := strconv.ParseBool(getStringAttribute(data, row, col))
	if v {
		return 1
	} else {
		return 0
	}
}

func parsePumpInstalled(v string) string {
	/*
		`1-1-014`
		First part (1- or 0-) indicates whether the pump is installed/fitted. If so (1- means it is)
		the second part indicates it's speed type.
		The third part represents it's possible states (0 OFF, 1 ON, 4 AUTO)
	*/
	return v
}

func parseRfResponse(data string) map[string][]string {
	/*
		RF:
		,R2,18,250,51,70,4,13,50,55,19,6,2020,376,9999,1,0,490,207,34,6000,602,23,20,0,0,0,0,44,35,45,:
		,R3,32,1,4,4,4,SW V5 17 05 31,SV3,18480001,20000826,1,0,0,0,0,0,NA,7,0,470,Filtering,4,0,7,7,0,0,:
		,R4,NORM,0,0,0,1,0,3547,4,20,4500,7413,567,1686,0,8388608,0,0,5,0,98,0,10084,4,80,100,0,0,4,:
		,R5,0,1,0,1,0,0,0,0,0,0,1,0,1,0,376,0,3,4,0,0,0,0,0,1,2,6,:
		,R6,1,5,30,2,5,8,1,360,1,0,3584,5120,127,128,5632,5632,2304,1792,0,30,0,0,0,0,2,3,0,:
		,R7,2304,0,1,1,1,0,1,0,0,0,253,191,253,240,483,125,77,1,0,0,0,23,200,1,0,1,31,32,35,100,5,:
		,R9,F1,255,0,0,0,0,0,0,0,0,0,0,:
		,RA,F2,0,0,0,0,0,0,255,0,0,0,0,:
		,RB,F3,0,0,0,0,0,0,0,0,0,0,0,:
		,RC,0,1,1,0,0,0,0,0,0,2,0,0,1,0,:
		,RE,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,-4,13,30,8,5,1,0,0,0,0,0,:*
		,RG,1,1,1,1,1,1,1-1-014,1-1-01,1-1-01,0-,0-,0,:*
	*/
	rows := strings.Split(data, "\n")
	ret := map[string][]string{}
	lines := []string{
		"R2",
		"R3",
		"R4",
		"R5",
		"R6",
		"R7",
		// No 8...
		"R9",
		"RA",
		"RB",
		"RC",
		// No D...
		"RE",
		"RG",
	}

	for _, rp := range lines {
		for _, row := range rows {
			if strings.HasPrefix(row, ","+rp) {
				ret[rp] = strings.Split(row[4:len(row)-2], ",")
				//fmt.Println(fmt.Sprintf("Row  %s: %s", rp, row))
			}
		}
	}

	return ret
}

func (s *SpaConn) Read() (SpaAttributes, error) {
	var attributes SpaAttributes

	//fmt.Println("Reading data from SPA")
	// Request RF data
	s.setTimeout(5)
	_, err := s.Conn.Write([]byte("RF\n"))
	if err != nil {
		s.Conn.Close()
		s.Conn = nil
		return attributes, err
	}

	// Read response from spa
	replyBytes := make([]byte, 1024)
	_, err = s.Conn.Read(replyBytes)
	if err != nil {
		s.Conn.Close()
		s.Conn = nil
		return attributes, err
	}
	s.setTimeout(0)

	reply := string(replyBytes)
	if !strings.Contains(reply, "RF:") {
		return attributes, errors.New("Malformed Response")
	}
	/*
		fmt.Println(reply)
		fmt.Println("-----")
	*/

	data := parseRfResponse(reply)

	attributes.WaterTemperature = getFloatAttribute(data, "R5", 14) / 10
	attributes.TargetTemperature = getFloatAttribute(data, "R6", 7) / 10
	attributes.Heating = getBoolAttribute(data, "R5", 11)
	attributes.Auto = getBoolAttribute(data, "R5", 11)
	attributes.Sanitising = getBoolAttribute(data, "R5", 15)
	attributes.Cleaning = getBoolAttribute(data, "R5", 10)
	attributes.Sleeping = getBoolAttribute(data, "R5", 9)

	attributes.Blower.Mode = getFloatAttribute(data, "RC", 9)
	attributes.Blower.Speed = getFloatAttribute(data, "R6", 0)

	attributes.Lights.Active = getBoolAttribute(data, "R5", 13)
	attributes.Lights.Mode = getFloatAttribute(data, "R6", 3)
	attributes.Lights.Brightness = getFloatAttribute(data, "R6", 1)
	attributes.Lights.Speed = getFloatAttribute(data, "R6", 4)
	attributes.Lights.Colour = getFloatAttribute(data, "R6", 1)

	attributes.Pumps = [5]SpaPumpAttributes{}
	attributes.Pumps[0].Id = 1
	attributes.Pumps[0].Installed = getStringAttribute(data, "RG", 6)
	attributes.Pumps[0].Active = getBoolAttribute(data, "R5", 17)

	attributes.Pumps[1].Id = 2
	attributes.Pumps[1].Installed = getStringAttribute(data, "RG", 7)
	attributes.Pumps[1].Active = getBoolAttribute(data, "R5", 19)
	attributes.Pumps[1].Ok = getBoolAttribute(data, "RG", 1)

	attributes.Pumps[2].Id = 3
	attributes.Pumps[2].Installed = getStringAttribute(data, "RG", 8)
	attributes.Pumps[2].Active = getBoolAttribute(data, "R5", 19)
	attributes.Pumps[2].Ok = getBoolAttribute(data, "RG", 2)

	attributes.Pumps[3].Id = 4
	attributes.Pumps[3].Installed = getStringAttribute(data, "RG", 9)
	attributes.Pumps[3].Active = getBoolAttribute(data, "R5", 20)
	attributes.Pumps[3].Ok = getBoolAttribute(data, "RG", 3)

	attributes.Pumps[4].Id = 5
	attributes.Pumps[4].Installed = getStringAttribute(data, "RG", 10)
	attributes.Pumps[4].Active = getBoolAttribute(data, "R5", 21)
	attributes.Pumps[4].Ok = getBoolAttribute(data, "RG", 4)

	// TODO: parse pump installed
	/*
		fmt.Printf("Pump 1 Installed: %s \n", attributes.Pumps[0].Installed)
		fmt.Printf("Pump 2 Installed: %s \n", attributes.Pumps[1].Installed)
		fmt.Printf("Pump 3 Installed: %s \n", attributes.Pumps[2].Installed)
		fmt.Printf("Pump 4 Installed: %s \n", attributes.Pumps[3].Installed)
		fmt.Printf("Pump 5 Installed: %s \n", attributes.Pumps[4].Installed)
	*/

	attributes.Settings = SpaSettings{
		Lock: getFloatAttribute(data, "RG", 11),
	}

	return attributes, nil
}
