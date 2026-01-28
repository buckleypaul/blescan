package ble

import "fmt"

// manufacturers maps Bluetooth SIG Company Identifiers to company names
// Source: https://www.bluetooth.com/specifications/assigned-numbers/company-identifiers/
var manufacturers = map[uint16]string{
	0x0001: "Nokia Mobile Phones",
	0x0002: "Intel Corp.",
	0x0003: "IBM Corp.",
	0x0004: "Toshiba Corp.",
	0x0006: "Microsoft",
	0x000D: "Texas Instruments Inc.",
	0x000F: "Broadcom Corporation",
	0x0010: "Qualcomm",
	0x0012: "Motorola",
	0x001D: "Qualcomm Technologies International, Ltd.",
	0x0025: "NXP Semiconductors",
	0x0030: "ST Microelectronics",
	0x0046: "MediaTek, Inc.",
	0x004C: "Apple, Inc.",
	0x0057: "Harman International Industries, Inc.",
	0x0059: "Nordic Semiconductor ASA",
	0x005D: "Realtek Semiconductor Corporation",
	0x0075: "Samsung Electronics Co. Ltd.",
	0x0078: "Nike, Inc.",
	0x0087: "Garmin International, Inc.",
	0x008A: "AAMP of America",
	0x008C: "BDE Technology Co., Ltd.",
	0x0094: "Beats Electronics",
	0x009E: "Bose Corporation",
	0x00D2: "Dialog Semiconductor B.V.",
	0x00E0: "Google",
	0x00EF: "Suunto Oy",
	0x0106: "Jawbone",
	0x010F: "Philips Lighting B.V.",
	0x0131: "Cypress Semiconductor Corporation",
	0x0154: "Huawei Technologies Co., Ltd.",
	0x0157: "Xiaomi Inc.",
	0x015D: "Polar Electro Oy",
	0x0171: "Amazon.com Services, Inc.",
	0x0180: "Anhui Huami Information Technology Co., Ltd.",
	0x018E: "Shenzhen Goodix Technology Co., Ltd.",
	0x0197: "SteelSeries ApS",
	0x01B7: "Facebook, Inc.",
	0x01C3: "Withings",
	0x01D7: "LEGO System A/S",
	0x01DA: "Murata Manufacturing Co., Ltd.",
	0x0203: "Amazfit",
	0x0224: "SAMSUNG ELECTRONICS CO., LTD.",
	0x022B: "Bragi GmbH",
	0x022D: "SmartThings, Inc.",
	0x0235: "Nothing Technology Limited",
	0x024F: "Espressif Incorporated",
	0x025A: "Ember Technologies, Inc.",
	0x026B: "Logitech International SA",
	0x028A: "Blue Yonder Group, Inc.",
	0x02A5: "DTS, Inc.",
	0x02B3: "Meta Platforms Technologies, LLC",
	0x02E1: "Fitbit, Inc.",
	0x02FD: "Skullcandy, Inc.",
	0x0310: "Tile, Inc.",
	0x031B: "Oura Health Oy",
	0x0339: "Sonos, Inc.",
	0x0362: "JBL",
	0x038F: "Xiaomi Communications Co., Ltd.",
	0x039A: "LG Electronics",
	0x03C3: "Peloton Interactive, Inc.",
	0x03DA: "WHOOP, Inc.",
	0x03E1: "Belkin International Inc.",
	0x0408: "OnePlus Electronics (Shenzhen) Co., Ltd.",
	0x041A: "Brilliant Home Technology, Inc.",
	0x042B: "Samsung Electronics Co., Ltd.",
	0x044E: "Govee Moments, LLC",
	0x0499: "Ruuvi Innovations Ltd.",
	0x057A: "Shenzhen Tuya Smart Technology Co., Ltd.",
	0x05A7: "Arlo Technologies, Inc.",
	0x0618: "eufy",
	0x0822: "OPPO",
	0x09A2: "Anker Innovations Limited",
	0x09FC: "Nothing (Shenzhen) Technology Co., Ltd.",
}

// GetManufacturerName returns the manufacturer name for a company ID
func GetManufacturerName(companyID uint16) string {
	if name, ok := manufacturers[companyID]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (0x%04X)", companyID)
}

// GetManufacturerNameWithID returns the manufacturer name with company ID
func GetManufacturerNameWithID(companyID uint16) string {
	if name, ok := manufacturers[companyID]; ok {
		return fmt.Sprintf("%s (0x%04X)", name, companyID)
	}
	return fmt.Sprintf("Unknown (0x%04X)", companyID)
}
