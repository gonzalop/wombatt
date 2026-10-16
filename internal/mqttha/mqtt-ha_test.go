package mqttha

import "testing"

func TestKey(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"battery_voltage", "battery_voltage"},
		{"GPIO13", "GPIO13"},
		{"PV1 Voltage", "PV1_Voltage"},
		{"DC/DC Transformer Temperature", "DC_DC_Transformer_Temperature"},
		{"Grid External Limiter 1 Power (CT1)", "Grid_External_Limiter_1_Power_CT1"},
	}
	for _, tt := range tests {
		if got := Key(tt.name); got != tt.want {
			t.Errorf("Key(%q) = %q; want %q", tt.name, got, tt.want)
		}
	}
}

func TestDiscoveryConfig(t *testing.T) {
	info := map[string]string{"name": "PV1 Voltage", "unit": "V", "precision": "1"}
	config := DiscoveryConfig(info, "ha/sensor/x_info/state", "x_PV1_Voltage", nil)
	checks := map[string]any{
		"name":                        "PV1 Voltage",
		"value_template":              "{{ value_json['PV1 Voltage'] }}",
		"unique_id":                   "x_PV1_Voltage",
		"default_entity_id":           "sensor.x_pv1_voltage",
		"unit_of_measurement":         "V",
		"suggested_display_precision": 1,
	}
	for k, want := range checks {
		if config[k] != want {
			t.Errorf("config[%q] = %v; want %v", k, config[k], want)
		}
	}
}
