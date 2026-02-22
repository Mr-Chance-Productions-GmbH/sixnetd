package zerotier

// ModeFlags returns the allowManaged/allowDNS/allowGlobal/allowDefault values
// for the requested connection mode.
func ModeFlags(mode NetworkMode) map[string]bool {
	switch mode {
	case ModeVPN:
		return map[string]bool{
			"allowManaged": true,
			"allowDNS":     true,
			"allowGlobal":  false,
			"allowDefault": false,
		}
	case ModeLAN:
		return map[string]bool{
			"allowManaged": true,
			"allowDNS":     true,
			"allowGlobal":  true,
			"allowDefault": false,
		}
	case ModeExit:
		return map[string]bool{
			"allowManaged": true,
			"allowDNS":     true,
			"allowGlobal":  true,
			"allowDefault": true,
		}
	default: // ModeDisconnected — clear everything
		return map[string]bool{
			"allowManaged": false,
			"allowDNS":     false,
			"allowGlobal":  false,
			"allowDefault": false,
		}
	}
}
