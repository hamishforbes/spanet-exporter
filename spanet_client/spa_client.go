package spanet_client

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type SpaPumpAttributes struct {
	Id        int
	Active    float64
	Installed float64
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
	socket    SpanetSocket
	tcpSocket *net.TCPConn
}

func (s *SpaConn) Connect() error {

	//fmt.Println("Connecting to spa at: " + s.socket.SpaUrl)
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.socket.SpaUrl)
	if err != nil {
		fmt.Println("Failed to resolve spa URL")
		return err
	}

	client, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Failed to connect to spa")
		return err
	}

	fmt.Println("Opened TCP socket to spa")

	connString := fmt.Sprintf("<connect--%d--%d>", s.socket.SocketId, s.socket.MemberId)
	//fmt.Println(connString)

	_, err = client.Write([]byte(connString))
	if err != nil {
		client.Close()
		fmt.Println("Failed to send data to spa")
		return err
	}

	// Check that data sent successfully
	reply := make([]byte, 22)
	client.Read(reply)
	if string(reply) != "Successfully connected" {
		// The spa has successfully connected
		return errors.New("Failed to handshake with spa" + string(reply))
	}

	//fmt.Println("Successfully connected to spa, ready to send/recieve commands")
	s.tcpSocket = client

	return nil
}

func getFloatAttribute(data [][]string, row int, col int) float64 {
	if len(data) < row {
		return 0
	}
	if len(data[row]) < col {
		return 0
	}
	v, _ := strconv.ParseFloat(data[row][col], 64)
	return v
}

func getBoolAttribute(data [][]string, row int, col int) float64 {
	if len(data) < row {
		return 0
	}
	if len(data[row]) < col {
		return 0
	}
	v, _ := strconv.ParseBool(data[row][col])
	if v {
		return 1
	} else {
		return 0
	}
}

func (s *SpaConn) Read() (SpaAttributes, error) {
	var attributes SpaAttributes
	//fmt.Println("Reading data from SPA")
	// Request RF data
	s.tcpSocket.Write([]byte("RF\n"))

	// Read response from spa
	replyBytes := make([]byte, 1024)
	s.tcpSocket.Read(replyBytes)

	reply := string(replyBytes)
	if !strings.Contains(reply, "RF:") {
		return attributes, errors.New("Malformed Response")
	}

	//fmt.Println(reply)
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

	data := [][]string{}
	for i, r := range strings.Split(reply, "\n") {
		if i == 0 {
			continue
		}
		//fmt.Println(fmt.Sprintf("Row %d: %s", i-1, r[4:len(r)-2]))
		var row = []string{}
		for _, v := range strings.Split(r[4:len(r)-2], ",") {
			row = append(row, v)
		}
		data = append(data, row)
	}

	attributes.WaterTemperature = getFloatAttribute(data, 3, 14) / 10
	attributes.TargetTemperature = getFloatAttribute(data, 4, 7) / 10
	attributes.Heating = getBoolAttribute(data, 3, 11)
	attributes.Auto = getBoolAttribute(data, 3, 11)
	attributes.Sanitising = getBoolAttribute(data, 3, 15)
	attributes.Cleaning = getBoolAttribute(data, 3, 10)
	attributes.Sleeping = getBoolAttribute(data, 3, 9)

	attributes.Blower.Mode = getFloatAttribute(data, 10, 9)
	attributes.Blower.Speed = getFloatAttribute(data, 4, 0)

	attributes.Lights.Active = getBoolAttribute(data, 3, 13)
	attributes.Lights.Mode = getFloatAttribute(data, 4, 3)
	attributes.Lights.Brightness = getFloatAttribute(data, 4, 1)
	attributes.Lights.Speed = getFloatAttribute(data, 4, 4)
	attributes.Lights.Colour = getFloatAttribute(data, 4, 1)

	attributes.Pumps = [5]SpaPumpAttributes{}
	attributes.Pumps[0].Id = 1
	attributes.Pumps[0].Installed = getBoolAttribute(data, 11, 6)
	attributes.Pumps[0].Active = getBoolAttribute(data, 3, 17)

	attributes.Pumps[1].Id = 2
	attributes.Pumps[1].Installed = getBoolAttribute(data, 11, 7)
	attributes.Pumps[1].Active = getBoolAttribute(data, 3, 19)
	attributes.Pumps[1].Ok = getBoolAttribute(data, 11, 1)

	attributes.Pumps[2].Id = 3
	attributes.Pumps[2].Installed = getBoolAttribute(data, 11, 8)
	attributes.Pumps[2].Active = getBoolAttribute(data, 3, 19)
	attributes.Pumps[2].Ok = getBoolAttribute(data, 11, 2)

	attributes.Pumps[3].Id = 4
	attributes.Pumps[3].Installed = getBoolAttribute(data, 11, 9)
	attributes.Pumps[3].Active = getBoolAttribute(data, 3, 20)
	attributes.Pumps[3].Ok = getBoolAttribute(data, 11, 3)

	attributes.Pumps[4].Id = 5
	attributes.Pumps[4].Installed = getBoolAttribute(data, 11, 10)
	attributes.Pumps[4].Active = getBoolAttribute(data, 3, 21)
	attributes.Pumps[4].Ok = getBoolAttribute(data, 11, 4)

	attributes.Settings = SpaSettings{
		Lock: getFloatAttribute(data, 11, 11),
	}

	return attributes, nil
}
