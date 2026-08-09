package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

const (
	comment = "//"
	define  = "#define"
)

func main() {
	data, err := ioutil.ReadFile("/opt/picoscope/include/libps2000a/PicoStatus.h")
	if err != nil {
		log.Println(err)
		return
	}
	lines := strings.Split(string(data), "\n")
	
	fmt.Println("package psc")
	fmt.Println("// Auto-generated from PicoStatus.h. Do not edit manually.")
	fmt.Println("// This file contains pure Go constants decoupled from any specific PicoScope driver.")
	fmt.Println("\nconst (")
	
	for _, line := range lines {
		if bytes.Contains([]byte(line), []byte(define)) {
			fields := bytes.Fields([]byte(line))
			if len(fields) >= 3 {
				name := string(fields[1])
				if strings.HasPrefix(name, "PICO_") {
					valStr := string(fields[2])
					// Remove "UL" or "L" from the end of the hex string if present
					valStr = strings.TrimSuffix(valStr, "UL")
					valStr = strings.TrimSuffix(valStr, "L")
					fmt.Printf("\t%s = %s\n", name, valStr)
				}
			}
		}
	}
	
	// Add some specific aliases from the old file
	fmt.Println("\n\tPicoDriverVersion             = PICO_DRIVER_VERSION")
	fmt.Println("\tPicoUsbVersion                = PICO_USB_VERSION")
	fmt.Println("\tPicoHardwareVersion           = PICO_HARDWARE_VERSION")
	fmt.Println("\tPicoVariantIfo                = PICO_VARIANT_INFO")
	fmt.Println("\tPicoBatchAndSerial            = PICO_BATCH_AND_SERIAL")
	fmt.Println("\tPicoCalDate                   = PICO_CAL_DATE")
	fmt.Println("\tPicoKernelVersion             = PICO_KERNEL_VERSION")
	fmt.Println("\tPicoDigitalHardwareVersion    = PICO_DIGITAL_HARDWARE_VERSION")
	fmt.Println("\tPicoAnaloguelHardwareVersion  = PICO_ANALOGUE_HARDWARE_VERSION")
	fmt.Println("\tPicoFirmwareVersion1          = PICO_FIRMWARE_VERSION_1")
	fmt.Println("\tPicoFirmwareVersion2          = PICO_FIRMWARE_VERSION_2")
	fmt.Println("\tPicoFirmwareVersion3          = PICO_FIRMWARE_VERSION_3")
	fmt.Println("\tPicoMacAddress                = PICO_MAC_ADDRESS")
	fmt.Println("\tPicoShadowCal                 = PICO_SHADOW_CAL")
	fmt.Println("\tPicoIppVersion                = PICO_IPP_VERSION")
	fmt.Println("\tPicoDriverPath                = PICO_DRIVER_PATH")
	fmt.Println("\tPicoFrontPanelFirmwareVersion = PICO_FRONT_PANEL_FIRMWARE_VERSION")
	fmt.Println("\tPicoBootloaderVersion         = PICO_BOOTLOADER_VERSION")
	fmt.Println("\tPicoOk                        = PICO_OK")
	
	fmt.Println(")")
	fmt.Println("\nvar statMap = map[int]string{")
	
	for i, line := range lines {
		if bytes.Contains([]byte(line), []byte(define)) {
			fields := bytes.Fields([]byte(line))
			if len(fields) >= 3 {
				name := string(fields[1])
				if strings.HasPrefix(name, "PICO_") {
					if i > 0 && bytes.Contains([]byte(lines[i-1]), []byte(comment)) {
						j := i - 1
						for j >= 0 && bytes.Contains([]byte(lines[j]), []byte(comment)) {
							j--
						}
						fmt.Printf("\t%s: \"%s ", name, name)
						for k := j + 1; k < i; k++ {
							cmt := bytes.TrimSpace(bytes.TrimPrefix(bytes.TrimSpace([]byte(lines[k])), []byte("//")))
							if len(cmt) > 0 {
								fmt.Printf("%s ", string(cmt))
							}
						}
						fmt.Printf("\",\n")
					} else {
						fmt.Printf("\t%s: \"%s\",\n", name, name)
					}
				}
			}
		}
	}
	
	fmt.Println("}")
	fmt.Println("\nfunc StatStr(code int) string {")
	fmt.Println("\treturn statMap[code]")
	fmt.Println("}")
}
