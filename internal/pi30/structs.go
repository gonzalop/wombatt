package pi30

// ResponseChecker is implemented for structs that were successfully read but
// contain what looks like invalid information.
type ResponseChecker interface {
	Valid() bool
}

type EmptyResponse struct {
	AckOrNak string `desc:"response"`
}

type Q1Response struct {
	EndOfAbsorbCharging    int16 `name:"end_of_absorb_charging" desc:"Time until the end of absorb charging" unit:"s" icon:"mdi:clock-time-two-outline"`
	EndOfFloatCharging     int16 `name:"end_of_float_charging" desc:"Time until the end of float charging" unit:"s" icon:"mdi:clock-time-two-outline"`
	SccCommunicating       int8  `name:"scc_flags" desc:"SCC flags" values:"0:Not communicating,1:Powered and communicating"`
	_                      string
	_                      string
	SccPwmTemperature      int8 `name:"scc_pwm_temperature" desc:"SCC PWM temperature" unit:"°C"`
	InverterTemperature    int8 `name:"inverter_temperature" desc:"Inverter temperature" unit:"°C"`
	BatteryTemperature     int8 `name:"battery_temperature" desc:"Battery temperature" unit:"°C"`
	TransformerTemperature int8 `name:"transformer_temperature" desc:"Transformer temperature" unit:"°C"`
	GPIO13                 int8 `name:"GPIO13"`
	FanLockStatus          int8 `name:"fan_lock_status" desc:"Fan lock status" values:"0:not locked,1:locked"`
	_                      string
	FanPwmSpeed            int8    `name:"fan_pwm_speed" desc:"Fan PWM speed" unit:"%"`
	SccChargePower         int16   `name:"scc_charge_power" desc:"SCC charge power" unit:"W" icon:"mdi:solar-power"`
	ParallelWarning        int8    `name:"parallel_warning" desc:"Parallel warning"`
	SyncFrequency          float32 `name:"sync_frequency" desc:"Sync frequency" unit:"Hz"`
	InverterChargeStatus   int8    `name:"inverter_charge_status" desc:"Inverter charger status" values:"10:not charging,11:bulk stage,12:absorb,13:float"`
	// Rest of fields in the response are ignored
}

type QPIGSResponse struct {
	GridVoltage                 float32 `name:"grid_voltage" desc:"Grid voltage" unit:"V"`
	GridFrequency               float32 `name:"grid_frequency" desc:"Grid frequency" unit:"Hz"`
	ACOutputVoltage             float32 `name:"ac_output_voltage" desc:"AC output voltage" unit:"V"`
	ACOutputFrequency           float32 `name:"ac_output_frequency" desc:"AC output frequency" unit:"Hz"`
	AcOutputApparentPower       int16   `name:"ac_output_apparent_power" desc:"AC output apparent power" unit:"VA"`
	AcOutputActivePower         int16   `name:"ac_output_active_power" desc:"AC output active power" unit:"W"`
	OutputLoadPercentage        int8    `name:"output_load_percentage" desc:"output load percentage" unit:"%"`
	BusVoltage                  float32 `name:"bus_voltage" desc:"BUS voltage" unit:"V"`
	BatteryVoltage              float32 `name:"battery_voltage" desc:"Battery voltage" unit:"V"`
	BatteryChargingCurrent      int16   `name:"battery_charging_current" desc:"Battery charging current" unit:"A" icon:"mdi:current-dc"`
	BatteryCapacity             int8    `name:"battery_capacity" desc:"Battery capacity" unit:"%"`
	InverterHeatSinkTemperature int8    `name:"internal_heat_sink_temperature" desc:"Inverter heat sink temperature" unit:"°C"`
	PV1InputVoltage             float32 `name:"pv1_input_voltage" desc:"PV1 input voltage" unit:"V"`
	PV1InputCurrent             float32 `name:"pv1_input_current" desc:"PV1 input current" unit:"A" icon:"mdi:current-dc"`
	BatteryVoltageSCC           float32 `name:"battery_voltage_scc" desc:"Battery voltage from SCC" unit:"V"`
	BatteryDischargeCurrent     int16   `name:"battery_discharge_current" desc:"Battery discharge current" unit:"A" icon:"mdi:current-dc"`
	DeviceStatus                uint8   `name:"device_status" desc:"Device status" parseas:"binary" flags:"Add SBU priority version,configuration changed,SCC firmware updated,Load on,Battery voltage steady while charging,Charging,SCC charging,AC charging"`
	BatteryVoltageOffset        float32 `name:"battery_voltage_offset" desc:"Battery voltage offset for fans on" unit:"mV"` // TODO: 10mV
	PV1ChargingPower            int16   `name:"pv1_charging_power" desc:"PV1 charging power" unit:"W"`
	DeviceStatusFlags           int8    `name:"device_status_flags" desc:"Additional device status flags"`
	SolarFeedToGrid             int8    `name:"solar_feed_to_grid" desc:"Solar feed to grid" values:"0:normal,1:solar feed the grid"`
	// CountryRegulations          int8    `name:"country_regulations" desc:"Country regulations" values:"00:India,01:Germany,02:South America"`
	// SolarFeedToGridPower        int16   `name:"solar_feed_to_grid_power" desc:"Solar feed to grid power" unit:"W"`
}

func (q *QPIGSResponse) Valid() bool {
	// Some times one of two inverters has mostly zeroes in its QPIGS response
	return (q.GridVoltage != 0.0 && q.GridFrequency != 0.0) ||
		(q.ACOutputVoltage != 0.0 && q.ACOutputFrequency != 0.0)
}

type QPIGS2Response struct {
	PV2InputCurrent  float32 `name:"pv2_input_current" desc:"PV2 input current" unit:"A" icon:"mdi:current-dc"`
	PV2InputVoltage  float32 `name:"pv2_input_voltage" desc:"PV2 input voltage" unit:"V"`
	PV2ChargingPower int16   `name:"pv2_charging_power" desc:"PV2 charging power" unit:"W"`
}

type QPIRIResponse struct {
	GridRatingVoltage           float32 `name:"grid_rating_voltage" desc:"Grid rating voltage" unit:"V"`
	GridRatingCurrent           float32 `name:"grid_rating_current" desc:"Grid rating current" unit:"A" icon:"mdi:current-ac"`
	ACOutputRatingVoltage       float32 `name:"grid_rating_voltage" desc:"AC output rating voltage" unit:"V"`
	ACOutputRatingFrequency     float32 `name:"grid_rating_frequency" desc:"AC output rating frequency" unit:"Hz"`
	ACOutputRatingCurrent       float32 `name:"ac_output_rating_current" desc:"AC output rating current" unit:"A" icon:"mdi:current-ac"`
	AcOutputRatingApparentPower int16   `name:"ac_output_rating_apparent_power" desc:"AC output rating apparent power" unit:"VA"`
	AcOutputRatingActivePower   int16   `name:"ac_output_rating_active_power" desc:"AC output rating active power" unit:"W"`
	BatteryVoltage              float32 `name:"battery_voltage" desc:"Battery voltage" unit:"V"`
	BatteryRechargeVoltage      float32 `name:"battery_recharge_voltage" desc:"Battery recharge voltage" unit:"V"`
	BatteryUnderVoltage         float32 `name:"battery_under_voltage" desc:"Battery under voltage" unit:"V"`
	BatteryBulkVoltage          float32 `name:"battery_bulk_voltage" desc:"Battery bulk voltage" unit:"V"`
	BatteryFloatVoltage         float32 `name:"battery_float_voltage" desc:"Battery float voltage" unit:"V"`
	BatteryType                 int8    `name:"battery_type" desc:"Battery type" values:"0:AGM,1:Flooded,2:User,3:unknown,4:Pylontech,5:WECO,6:Soltaro,7:LIb-protocol compatible,8:3rd party lithium"`
	MaxACChargingCurrent        int16   `name:"max_ac_charging_current" desc:"Max AC charging current" unit:"A" icon:"mdi:current-ac"`
	MaxChargingCurrent          int16   `name:"max_charging_current" desc:"Max charging current" unit:"A" icon:"mdi:current-dc"`
	InputVoltageRange           int8    `name:"input_voltage_range" desc:"Input voltage range" values:"0:Appliance,1:UPS"`
	OutputSourcePriority        int8    `name:"output_source_priority" desc:"Output source priority" values:"0:USB,1:SUB,2:SBU"`
	ChargerSourcePriority       int8    `name:"charger_source_priority" desc:"Charger source priority" values:"1:Solar first,2:Solar + utility,3:Only solar"`
	ParalellMaxNum              int8    `name:"parallel_max_num" desc:"Parallel max num"`
	MachineType                 int8    `name:"machine_type" desc:"Machine type" values:"00:grid tie,01:off-grid,02:hybrid"`
	Topology                    int8    `name:"topology" desc:"Topology" values:"0:transformerless,1:transformer"`
	OutputMode                  int8    `name:"output_mode" desc:"Output mode" values:"0:Single machine,1:Parallel output,2:Phase 1 of 3 phase output,3:Phase 2 of 3 phase output,4:Phase 3 of 3 phase output,5:Phase 1 of 2 phase output,6:Phase 2 of 2 phase output (120°),7:Phase 2 of 2 phase output (180°)"`
	BatteryRedischargeVoltage   float32 `name:"battery_redischarge_voltage" desc:"Battery redischarge voltage" unit:"V"`
	PVOKCondition               int8    `name:"pv_ok_condition" desc:"PV OK condition for parallel" values:"0:one inverter connected to PV is enough,1:All inverters need to have PV for PV to be OK"`
	PVPowerBalance              int8    `name:"pv_power_balance" desc:"PV power balance" values:"0:PV input max current will be the max charged current,1: PV input max power will be the sum of the max charged power and loads power"`
	MaxChargingTimeAtCV         int16   `name:"max_charging_time_at_cv" desc:"Max charging time at C.V." unit:"m"`
	OperationLogic              int8    `name:"operation_logic" desc:"Operation logic" values:"0:Automatic,1:On-line,2:ECO"`
	MaxDischargingCurrent       int8    `name:"max_discharging_current" desc:"Max discharging current" unit:"A" icon:"mdi:current-dc"`
}

type QPGSResponse struct {
	Instance                   int     `name:"parallel_instance_number" desc:"Parallel instance number"`
	Serial                     string  `name:"serial_number" desc:"Serial number"`
	WorkMode                   string  `name:"work_mode" desc:"Work mode" values:"P:Power On,S:Standby,L:Line,B:Battery,F:Fault,H:Power Saving,D:Shutdown"`
	FaultCode                  int16   `name:"fault_code" desc:"Fault code"`
	GridVoltage                float32 `name:"grid_voltage" desc:"Grid voltage" unit:"V"`
	GridFrequency              float32 `name:"grid_frequency" desc:"Grid frequency" unit:"Hz"`
	ACOutputVoltage            float32 `name:"ac_output_voltage" desc:"AC output voltage" unit:"V"`
	ACOutputFrequency          float32 `name:"ac_output_fequency" desc:"AC output frequency" unit:"Hz"`
	AcOutputApparentPower      int16   `name:"ac_output_apparent_power" desc:"AC output apparent power" unit:"VA"`
	AcOutputActivePower        int16   `name:"ac_output_active_power" desc:"AC output active power" unit:"W"`
	LoadPercentage             int8    `name:"load_percentage" desc:"Load percentage" unit:"%"`
	BatteryVoltage             float32 `name:"battery_voltage" desc:"Battery voltage" unit:"V"`
	BatteryChargingCurrent     int16   `name:"battery_charging_current" desc:"Battery charging current" unit:"A" icon:"mdi:current-dc"`
	BatteryCapacity            int8    `name:"battery_capacity" desc:"Battery capacity" unit:"%"`
	PV1InputVoltage            float32 `name:"pv1_input_voltage" desc:"PV1 input voltage" unit:"V"`
	TotalChargingCurrent       int16   `name:"total_charging_current" desc:"Total charging current" unit:"A" icon:"mdi:current-dc"`
	TotalACOutputApparentPower int16   `name:"total_ac_output_apparent_power" desc:"Total AC output apparent power" unit:"VA"`
	TotalOutputActivePower     int16   `name:"total_output_active_power" desc:"Total output active power" unit:"W"`
	TotalACOutputPercentage    int8    `name:"total_ac_output_percentage" desc:"Total AC output percentage" unit:"%"`
	InverterStatus             string  `name:"inverter_status" desc:"Inverter status" bitgroups:"SCC OK|AC charging|SCC charging|Battery over voltage,Battery under voltage|Line loss|Load on|Configuration changed"`
	OutputMode                 int8    `name:"output_mode" desc:"Output mode" values:"0:Single machine,1:Parallel output,2:Phase 1 of 3 phase output,3:Phase 2 of 3 phase output,4:Phase 3 of 3 phase output,5:Phase 1 of 2 phase output,6:Phase 2 of 2 phase output (120°),7:Phase 2 of 2 phase output (180°)"`
	ChargerSourcePriority      int8    `name:"charger_source_priority" desc:"Charger source priority" values:"0:Utility first,1:Solar first,2:Solar + utility,3:Solar only"`
	MaxChargerCurrent          int16   `name:"max_charger_current" desc:"Max charger current" unit:"A" icon:"mdi:current-dc"`
	MaxChargingRange           int16   `name:"max_charging_range" desc:"Max charging range" unit:"A" icon:"mdi:current-dc"`
	MaxACChargerCurrent        int16   `name:"max_ac_charger_current" desc:"Max AC charger current" unit:"A" icon:"mdi:current-ac"`
	PV1InputCurrent            float32 `name:"pv1_input_current" desc:"PV1 input current" unit:"A" icon:"mdi:current-dc"`
	BatteryDischargeCurrent    int16   `name:"battery_discharge_current" desc:"Battery discharge current" unit:"A" icon:"mdi:current-dc"`
	PV2InputVoltage            float32 `name:"pv2_input_voltage" desc:"PV2 input voltage" unit:"V"`
	PV2InputCurrent            int8    `name:"pv2_input_current" desc:"PV2 input current" unit:"A" icon:"mdi:current-dc"`
}
