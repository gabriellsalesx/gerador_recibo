package pix

import "fmt"

func CRC16(payload string) string {
	var crc uint16 = 0xFFFF
	for _, b := range []byte(payload) {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc)
}
