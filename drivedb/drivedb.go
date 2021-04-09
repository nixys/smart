// Copyright 2017-18 Daniel Swarbrick. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drivedb

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// SMART attribute conversion rule
type AttrConv struct {
	Conv string `yaml:"conv"`
	Name string `yaml:"name"`
}

type DriveModel struct {
	Family         string              `yaml:"family"`
	ModelRegex     string              `yaml:"model_regex"`
	FirmwareRegex  string              `yaml:"firmware_regex"`
	WarningMsg     string              `yaml:"warning"`
	Presets        map[string]AttrConv `yaml:"presets"`
	CompiledRegexp *regexp.Regexp
}

type DriveDb struct {
	Drives []DriveModel `yaml:"drives"`
}

var DbYaml = `
drives:
- family: DEFAULT
  model_regex: '-'
  firmware_regex: '-'
  warning: Default settings
  presets:
    "1":
      conv: raw48
      name: Raw_Read_Error_Rate
    "2":
      conv: raw48
      name: Throughput_Performance
    "3":
      conv: raw16(avg16)
      name: Spin_Up_Time
    "4":
      conv: raw48
      name: Start_Stop_Count
    "5":
      conv: raw16(raw16)
      name: Reallocated_Sector_Ct
    "6":
      conv: raw48
      name: Read_Channel_Margin
    "7":
      conv: raw48
      name: Seek_Error_Rate
    "8":
      conv: raw48
      name: Seek_Time_Performance
    "9":
      conv: raw24(raw8)
      name: Power_On_Hours
    "10":
      conv: raw48
      name: Spin_Retry_Count
    "11":
      conv: raw48
      name: Calibration_Retry_Count
    "12":
      conv: raw48
      name: Power_Cycle_Count
    "13":
      conv: raw48
      name: Read_Soft_Error_Rate
    "175":
      conv: raw48
      name: Program_Fail_Count_Chip
    "176":
      conv: raw48
      name: Erase_Fail_Count_Chip
    "177":
      conv: raw48
      name: Wear_Leveling_Count
    "178":
      conv: raw48
      name: Used_Rsvd_Blk_Cnt_Chip
    "179":
      conv: raw48
      name: Used_Rsvd_Blk_Cnt_Tot
    "180":
      conv: raw48
      name: Unused_Rsvd_Blk_Cnt_Tot
    "181":
      conv: raw48
      name: Program_Fail_Cnt_Total
    "182":
      conv: raw48
      name: Erase_Fail_Count_Total
    "183":
      conv: raw48
      name: Runtime_Bad_Block
    "184":
      conv: raw48
      name: End-to-End_Error
    "187":
      conv: raw48
      name: Reported_Uncorrect
    "188":
      conv: raw48
      name: Command_Timeout
    "189":
      conv: raw48
      name: High_Fly_Writes
    "190":
      conv: tempminmax
      name: Airflow_Temperature_Cel
    "191":
      conv: raw48
      name: G-Sense_Error_Rate
    "192":
      conv: raw48
      name: Power-Off_Retract_Count
    "193":
      conv: raw48
      name: Load_Cycle_Count
    "194":
      conv: tempminmax
      name: Temperature_Celsius
    "195":
      conv: raw48
      name: Hardware_ECC_Recovered
    "196":
      conv: raw16(raw16)
      name: Reallocated_Event_Count
    "197":
      conv: raw48
      name: Current_Pending_Sector
    "198":
      conv: raw48
      name: Offline_Uncorrectable
    "199":
      conv: raw48
      name: UDMA_CRC_Error_Count
    "200":
      conv: raw48
      name: Multi_Zone_Error_Rate
    "201":
      conv: raw48
      name: Soft_Read_Error_Rate
    "202":
      conv: raw48
      name: Data_Address_Mark_Errs
    "203":
      conv: raw48
      name: Run_Out_Cancel
    "204":
      conv: raw48
      name: Soft_ECC_Correction
    "205":
      conv: raw48
      name: Thermal_Asperity_Rate
    "206":
      conv: raw48
      name: Flying_Height
    "207":
      conv: raw48
      name: Spin_High_Current
    "208":
      conv: raw48
      name: Spin_Buzz
    "209":
      conv: raw48
      name: Offline_Seek_Performnce
    "220":
      conv: raw48
      name: Disk_Shift
    "221":
      conv: raw48
      name: G-Sense_Error_Rate
    "222":
      conv: raw48
      name: Loaded_Hours
    "223":
      conv: raw48
      name: Load_Retry_Count
    "224":
      conv: raw48
      name: Load_Friction
    "225":
      conv: raw48
      name: Load_Cycle_Count
    "226":
      conv: raw48
      name: Load-in_Time
    "227":
      conv: raw48
      name: Torq-amp_Count
    "228":
      conv: raw48
      name: Power-off_Retract_Count
    "230":
      conv: raw48
      name: Head_Amplitude
    "231":
      conv: raw48
      name: Temperature_Celsius
    "232":
      conv: raw48
      name: Available_Reservd_Space
    "233":
      conv: raw48
      name: Media_Wearout_Indicator
    "240":
      conv: raw24(raw8)
      name: Head_Flying_Hours
    "241":
      conv: raw48
      name: Total_LBAs_Written
    "242":
      conv: raw48
      name: Total_LBAs_Read
    "250":
      conv: raw48
      name: Read_Error_Retry_Rate
    "254":
      conv: raw48
      name: Free_Fall_Sensor
`

// LookupDrive returns the most appropriate DriveModel for a given ATA IDENTIFY value.
func (db *DriveDb) LookupDrive(ident []byte) DriveModel {
	var model DriveModel

	for _, d := range db.Drives {
		// Skip placeholder entry
		if strings.HasPrefix(d.Family, "$Id") {
			continue
		}

		if d.Family == "DEFAULT" {
			model = d
			continue
		}

		if d.CompiledRegexp.Match(ident) {
			model.Family = d.Family
			model.ModelRegex = d.ModelRegex
			model.FirmwareRegex = d.FirmwareRegex
			model.WarningMsg = d.WarningMsg
			model.CompiledRegexp = d.CompiledRegexp

			for id, p := range d.Presets {
				if _, exists := model.Presets[id]; exists {
					// Some drives override the conv but don't specify a name, so copy it from default
					if p.Name == "" {
						p.Name = model.Presets[id].Name
					}
				}
				model.Presets[id] = AttrConv{Name: p.Name, Conv: p.Conv}
			}

			break
		}
	}

	return model
}

// OpenDriveDb opens a YAML-formatted drive database, unmarshalls it, and returns a DriveDb.
func OpenDriveDb(dbYaml string) (DriveDb, error) {
	var db DriveDb

	if dbYaml == "" {
		dbYaml = DbYaml
	}
	//
	//f, err := os.Open(dbFile)
	//if err != nil {
	//	return db, nil
	//}
	//
	//defer f.Close()
	//dec := yaml.NewDecoder(f)
	//
	//if err := dec.Decode(&db); err != nil {
	//	return db, err
	//}

	err := yaml.Unmarshal([]byte(dbYaml), &db)
	if err != nil {
		return db, err
	}

	for i, d := range db.Drives {
		db.Drives[i].CompiledRegexp, _ = regexp.Compile(d.ModelRegex)
	}

	return db, nil
}
