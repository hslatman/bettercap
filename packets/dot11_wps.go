package packets

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type wpsAttrType int

const (
	wpsHex wpsAttrType = 0
	wpsStr wpsAttrType = 1
)

type wpsAttr struct {
	Name string
	Type wpsAttrType
	Func func([]byte, *map[string]string) string
	Desc map[string]string
}

type wpsDevType struct {
	Category string
	Subcats  map[uint16]string
}

var (
	wpsSignatureBytes = []byte{0x00, 0x50, 0xf2, 0x04}
	wfaExtensionBytes = []byte{0x00, 0x37, 0x2a}
	wpsVersion2ID     = uint8(0x00)
	wpsVersionDesc    = map[string]string{
		"10": "1.0",
		"11": "1.1",
		"20": "2.0",
	}

	wpsDeviceTypes = map[uint16]wpsDevType{
		0x0001: wpsDevType{"Computer", map[uint16]string{
			0x0001: "PC",
			0x0002: "Server",
			0x0003: "Media Center",
		}},
		0x0002: wpsDevType{"Input Device", map[uint16]string{}},
		0x0003: wpsDevType{"Printers, Scanners, Faxes and Copiers", map[uint16]string{
			0x0001: "Printer",
			0x0002: "Scanner",
		}},
		0x0004: wpsDevType{"Camera", map[uint16]string{
			0x0001: "Digital Still Camera",
		}},
		0x0005: wpsDevType{"Storage", map[uint16]string{
			0x0001: "NAS",
		}},
		0x0006: wpsDevType{"Network Infra", map[uint16]string{
			0x0001: "AP",
			0x0002: "Router",
			0x0003: "Switch",
		}},

		0x0007: wpsDevType{"Display", map[uint16]string{
			0x0001: "TV",
			0x0002: "Electronic Picture Frame",
			0x0003: "Projector",
		}},

		0x0008: wpsDevType{"Multimedia Device", map[uint16]string{
			0x0001: "DAR",
			0x0002: "PVR",
			0x0003: "MCX",
		}},

		0x0009: wpsDevType{"Gaming Device", map[uint16]string{
			0x0001: "XBox",
			0x0002: "XBox360",
			0x0003: "Playstation",
		}},
		0x000F: wpsDevType{"Telephone", map[uint16]string{
			0x0001: "Windows Mobile",
		}},
	}

	wpsAttributes = map[uint16]wpsAttr{
		0x104A: wpsAttr{Name: "Version", Desc: wpsVersionDesc},
		0x1044: wpsAttr{Name: "State", Desc: map[string]string{
			"01": "Not Configured",
			"02": "Configured",
		}},
		0x1012: wpsAttr{Name: "Device Password ID", Desc: map[string]string{
			"0000": "Pin",
			"0004": "PushButton",
		}},
		0x103B: wpsAttr{Name: "Response Type", Desc: map[string]string{
			"00": "Enrollee Info",
			"01": "Enrollee",
			"02": "Registrar",
			"03": "AP",
		}},

		0x1054: wpsAttr{Name: "Primary Device Type", Func: dot11ParseWPSDeviceType},
		0x1049: wpsAttr{Name: "Vendor Extension", Func: dot11ParseWPSVendorExtension},
		0x1053: wpsAttr{Name: "Selected Registrar Config Methods", Func: dot11ParseWPSConfigMethods},
		0x1008: wpsAttr{Name: "Config Methods", Func: dot11ParseWPSConfigMethods},
		0x103C: wpsAttr{Name: "RF Bands", Func: dott11ParseWPSBands},

		0x1057: wpsAttr{Name: "AP Setup Locked"},
		0x1041: wpsAttr{Name: "Selected Registrar"},
		0x1047: wpsAttr{Name: "UUID-E"},
		0x1021: wpsAttr{Name: "Manufacturer", Type: wpsStr},
		0x1023: wpsAttr{Name: "Model Name", Type: wpsStr},
		0x1024: wpsAttr{Name: "Model Number", Type: wpsStr},
		0x1042: wpsAttr{Name: "Serial Number", Type: wpsStr},
		0x1011: wpsAttr{Name: "Device Name", Type: wpsStr},
		0x1045: wpsAttr{Name: "SSID", Type: wpsStr},
		0x102D: wpsAttr{Name: "OS Version", Type: wpsStr},
	}

	wpsConfigs = map[uint16]string{
		0x0001: "USB",
		0x0002: "Ethernet",
		0x0004: "Label",
		0x0008: "Display",
		0x0010: "External NFC",
		0x0020: "Internal NFC",
		0x0040: "NFC Interface",
		0x0080: "Push Button",
		0x0100: "Keypad",
	}

	wpsBands = map[uint8]string{
		0x01: "2.4Ghz",
		0x02: "5.0Ghz",
	}
)

func dott11ParseWPSBands(data []byte, info *map[string]string) string {
	if len(data) == 1 {
		mask := uint8(data[0])
		bands := []string{}

		for bit, band := range wpsBands {
			if mask&bit != 0 {
				bands = append(bands, band)
			}
		}

		if len(bands) > 0 {
			return strings.Join(bands, ", ")
		}
	}

	return hex.EncodeToString(data)
}

func dot11ParseWPSConfigMethods(data []byte, info *map[string]string) string {
	if len(data) == 2 {
		mask := binary.BigEndian.Uint16(data)
		configs := []string{}

		for bit, conf := range wpsConfigs {
			if mask&bit != 0 {
				configs = append(configs, conf)
			}
		}

		if len(configs) > 0 {
			return strings.Join(configs, ", ")
		}
	}

	return hex.EncodeToString(data)
}

func dot11ParseWPSVendorExtension(data []byte, info *map[string]string) string {
	if len(data) > 3 && bytes.Equal(data[0:3], wfaExtensionBytes) {
		size := len(data)
		for offset := 3; offset < size; {
			idByte := uint8(data[offset])
			sizeByte := uint8(data[offset+1])
			if idByte == wpsVersion2ID {
				verByte := fmt.Sprintf("%x", data[offset+2])
				(*info)["Version"] = wpsVersionDesc[verByte]
				data = data[offset+3:]
				break
			}
			offset += int(sizeByte) + 2
		}
	}
	return hex.EncodeToString(data)
}

func dot11ParseWPSDeviceType(data []byte, info *map[string]string) string {
	if len(data) == 8 {
		catId := binary.BigEndian.Uint16(data[0:2])
		oui := data[2:6]
		subCatId := binary.BigEndian.Uint16(data[6:8])
		if cat, found := wpsDeviceTypes[catId]; found {
			if sub, found := cat.Subcats[subCatId]; found {
				return fmt.Sprintf("%s (oui:%x)", sub, oui)
			}
			return fmt.Sprintf("%s (oui:%x)", cat.Category, oui)
		}
		return fmt.Sprintf("cat:%x sub:%x oui:%x %x", catId, subCatId, oui, data)
	}
	return hex.EncodeToString(data)
}

func wpsUint16At(data []byte, size int, offset *int) (bool, uint16) {
	if *offset <= size-2 {
		off := *offset
		*offset += 2
		return true, binary.BigEndian.Uint16(data[off:])
	}
	// fmt.Printf("uint16At( data(%d), off=%d )\n", size, *offset)
	return false, 0
}

func wpsDataAt(data []byte, size int, offset *int, num int) (bool, []byte) {
	max := size - num
	if *offset <= max {
		off := *offset
		*offset += num
		return true, data[off : off+num]
	}
	// fmt.Printf("dataAt( data(%d), off=%d, num=%d )\n", size, *offset, num)
	return false, nil
}

func dot11ParseWPSTag(id uint16, size uint16, data []byte, info *map[string]string) {
	name := ""
	val := ""

	if attr, found := wpsAttributes[id]; found {
		name = attr.Name
		if attr.Type == wpsStr {
			val = string(data)
		} else {
			val = hex.EncodeToString(data)
		}

		if attr.Desc != nil {
			if desc, found := attr.Desc[val]; found {
				val = desc
			}
		}

		if attr.Func != nil {
			val = attr.Func(data, info)
		}
	} else {
		name = fmt.Sprintf("0x%X", id)
		val = hex.EncodeToString(data)
	}

	if val != "" {
		(*info)[name] = val
	}
}

func dot11ParseWPSData(data []byte) (ok bool, info map[string]string) {
	info = map[string]string{}
	size := len(data)

	for offset := 0; offset < size; {
		ok := false
		tagId := uint16(0)
		tagLen := uint16(0)
		tagData := []byte(nil)

		if ok, tagId = wpsUint16At(data, size, &offset); !ok {
			break
		} else if ok, tagLen = wpsUint16At(data, size, &offset); !ok {
			break
		} else if ok, tagData = wpsDataAt(data, size, &offset, int(tagLen)); !ok {
			break
		} else {
			dot11ParseWPSTag(tagId, tagLen, tagData, &info)
		}
	}

	return true, info
}

func Dot11ParseWPS(packet gopacket.Packet, dot11 *layers.Dot11) (ok bool, bssid net.HardwareAddr, info map[string]string) {
	ok = false
	for _, layer := range packet.Layers() {
		if layer.LayerType() == layers.LayerTypeDot11InformationElement {
			if dot11info, infoOk := layer.(*layers.Dot11InformationElement); infoOk && dot11info.ID == layers.Dot11InformationElementIDVendor {
				if bytes.Equal(dot11info.OUI, wpsSignatureBytes) {
					bssid = dot11.Address3
					ok, info = dot11ParseWPSData(dot11info.Info)
					return
				}
			}
		}
	}
	return
}
