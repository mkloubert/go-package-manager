package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mkloubert/go-package-manager/utils"
)

// SettingsFile handles a settings.yaml file
type SettingsFile struct {
	app  *AppContext
	data map[string]interface{}
}

// sf.GetFloat32() - returns a string value from settings via dot-notation
func (sf *SettingsFile) GetFloat32(name string, flagValue float32, defaultValue float32, options ...GetSettingOptions) float32 {
	return sf.getValue(
		name,
		flagValue, defaultValue,
		func(input interface{}, defaultValue interface{}) interface{} {
			s := strings.TrimSpace(
				fmt.Sprintf("%v", input),
			)

			f, err := strconv.ParseFloat(s, 32)
			if err == nil {
				return float32(f)
			}
			return defaultValue
		},
		options...,
	).(float32)
}

// sf.GetString() - returns a string value from settings via dot-notation
func (sf *SettingsFile) GetString(name string, flagValue string, defaultValue string, options ...GetSettingOptions) string {
	return sf.getValue(
		name,
		flagValue, defaultValue,
		func(input interface{}, defaultValue interface{}) interface{} {
			s := input.(string)
			if s == "" {
				return defaultValue
			}
			return s
		},
		options...,
	).(string)
}

func (sf *SettingsFile) getValue(
	name string,
	flagValue interface{}, defaultValue interface{},
	convertValue func(interface{}, interface{}) interface{},
	options ...GetSettingOptions,
) interface{} {
	symbolValue := &struct{}{}

	name = strings.TrimSpace(
		strings.ToLower(name),
	)

	// collect options
	doNotTrimEnvValues := false
	for _, o := range options {
		if o.DoNotTrimEnvValues != nil {
			doNotTrimEnvValues = *o.DoNotTrimEnvValues
		}
	}

	var value interface{} = flagValue

	// first try flag value
	if value == defaultValue {
		// no => try environment variable
		// foo.bar => GPM_FOO_BAR
		envName := "GPM_" + strings.TrimSpace(
			strings.ToUpper(
				strings.ReplaceAll(name, ".", "_"),
			),
		)

		envValue := sf.app.GetEnvValue(envName)
		if !doNotTrimEnvValues {
			envValue = strings.TrimSpace(envValue)
		}

		value = convertValue(envValue, defaultValue)
		if value == defaultValue {
			// now try if there is a setting in "sections" of gpm.yaml file
			gpmFileSettings := sf.app.GpmFile.GetSettingsSectionByEnvSafe(
				sf.app.GetEnvironment(),
			)

			settingsValue, err := utils.GetValueFromMap(gpmFileSettings, name, symbolValue)
			if err == nil && settingsValue != symbolValue {
				// yes
				value = settingsValue
			} else {
				// no => now finally try global settings.yaml file

				globalSettingsValue, err := utils.GetValueFromMap(sf.data, name, symbolValue)
				if err == nil && globalSettingsValue != symbolValue {
					value = globalSettingsValue
				}
			}
		}
	}

	return convertValue(value, defaultValue)
}
