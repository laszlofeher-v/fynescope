package psc

// Auto-generated from PicoStatus.h. Do not edit manually.
// This file contains pure Go constants decoupled from any specific PicoScope driver.

const (
	PICO_DRIVER_VERSION                                                          = 0x00000000
	PICO_USB_VERSION                                                             = 0x00000001
	PICO_HARDWARE_VERSION                                                        = 0x00000002
	PICO_VARIANT_INFO                                                            = 0x00000003
	PICO_BATCH_AND_SERIAL                                                        = 0x00000004
	PICO_CAL_DATE                                                                = 0x00000005
	PICO_KERNEL_VERSION                                                          = 0x00000006
	PICO_DIGITAL_HARDWARE_VERSION                                                = 0x00000007
	PICO_ANALOGUE_HARDWARE_VERSION                                               = 0x00000008
	PICO_FIRMWARE_VERSION_1                                                      = 0x00000009
	PICO_FIRMWARE_VERSION_2                                                      = 0x0000000A
	PICO_MAC_ADDRESS                                                             = 0x0000000B
	PICO_SHADOW_CAL                                                              = 0x0000000C
	PICO_IPP_VERSION                                                             = 0x0000000D
	PICO_DRIVER_PATH                                                             = 0x0000000E
	PICO_FIRMWARE_VERSION_3                                                      = 0x0000000F
	PICO_FRONT_PANEL_FIRMWARE_VERSION                                            = 0x00000010
	PICO_BOOTLOADER_VERSION                                                      = 0x10000001
	PICO_OK                                                                      = 0x00000000
	PICO_MAX_UNITS_OPENED                                                        = 0x00000001
	PICO_MEMORY_FAIL                                                             = 0x00000002
	PICO_NOT_FOUND                                                               = 0x00000003
	PICO_FW_FAIL                                                                 = 0x00000004
	PICO_OPEN_OPERATION_IN_PROGRESS                                              = 0x00000005
	PICO_OPERATION_FAILED                                                        = 0x00000006
	PICO_NOT_RESPONDING                                                          = 0x00000007
	PICO_CONFIG_FAIL                                                             = 0x00000008
	PICO_KERNEL_DRIVER_TOO_OLD                                                   = 0x00000009
	PICO_EEPROM_CORRUPT                                                          = 0x0000000A
	PICO_OS_NOT_SUPPORTED                                                        = 0x0000000B
	PICO_INVALID_HANDLE                                                          = 0x0000000C
	PICO_INVALID_PARAMETER                                                       = 0x0000000D
	PICO_INVALID_TIMEBASE                                                        = 0x0000000E
	PICO_INVALID_VOLTAGE_RANGE                                                   = 0x0000000F
	PICO_INVALID_CHANNEL                                                         = 0x00000010
	PICO_INVALID_TRIGGER_CHANNEL                                                 = 0x00000011
	PICO_INVALID_CONDITION_CHANNEL                                               = 0x00000012
	PICO_NO_SIGNAL_GENERATOR                                                     = 0x00000013
	PICO_STREAMING_FAILED                                                        = 0x00000014
	PICO_BLOCK_MODE_FAILED                                                       = 0x00000015
	PICO_NULL_PARAMETER                                                          = 0x00000016
	PICO_ETS_MODE_SET                                                            = 0x00000017
	PICO_DATA_NOT_AVAILABLE                                                      = 0x00000018
	PICO_STRING_BUFFER_TO_SMALL                                                  = 0x00000019
	PICO_ETS_NOT_SUPPORTED                                                       = 0x0000001A
	PICO_AUTO_TRIGGER_TIME_TO_SHORT                                              = 0x0000001B
	PICO_BUFFER_STALL                                                            = 0x0000001C
	PICO_TOO_MANY_SAMPLES                                                        = 0x0000001D
	PICO_TOO_MANY_SEGMENTS                                                       = 0x0000001E
	PICO_PULSE_WIDTH_QUALIFIER                                                   = 0x0000001F
	PICO_DELAY                                                                   = 0x00000020
	PICO_SOURCE_DETAILS                                                          = 0x00000021
	PICO_CONDITIONS                                                              = 0x00000022
	PICO_USER_CALLBACK                                                           = 0x00000023
	PICO_DEVICE_SAMPLING                                                         = 0x00000024
	PICO_NO_SAMPLES_AVAILABLE                                                    = 0x00000025
	PICO_SEGMENT_OUT_OF_RANGE                                                    = 0x00000026
	PICO_BUSY                                                                    = 0x00000027
	PICO_STARTINDEX_INVALID                                                      = 0x00000028
	PICO_INVALID_INFO                                                            = 0x00000029
	PICO_INFO_UNAVAILABLE                                                        = 0x0000002A
	PICO_INVALID_SAMPLE_INTERVAL                                                 = 0x0000002B
	PICO_TRIGGER_ERROR                                                           = 0x0000002C
	PICO_MEMORY                                                                  = 0x0000002D
	PICO_SIG_GEN_PARAM                                                           = 0x0000002E
	PICO_SHOTS_SWEEPS_WARNING                                                    = 0x0000002F
	PICO_SIGGEN_TRIGGER_SOURCE                                                   = 0x00000030
	PICO_AUX_OUTPUT_CONFLICT                                                     = 0x00000031
	PICO_AUX_OUTPUT_ETS_CONFLICT                                                 = 0x00000032
	PICO_WARNING_EXT_THRESHOLD_CONFLICT                                          = 0x00000033
	PICO_WARNING_AUX_OUTPUT_CONFLICT                                             = 0x00000034
	PICO_SIGGEN_OUTPUT_OVER_VOLTAGE                                              = 0x00000035
	PICO_DELAY_NULL                                                              = 0x00000036
	PICO_INVALID_BUFFER                                                          = 0x00000037
	PICO_SIGGEN_OFFSET_VOLTAGE                                                   = 0x00000038
	PICO_SIGGEN_PK_TO_PK                                                         = 0x00000039
	PICO_CANCELLED                                                               = 0x0000003A
	PICO_SEGMENT_NOT_USED                                                        = 0x0000003B
	PICO_INVALID_CALL                                                            = 0x0000003C
	PICO_GET_VALUES_INTERRUPTED                                                  = 0x0000003D
	PICO_NOT_USED                                                                = 0x0000003F
	PICO_INVALID_SAMPLERATIO                                                     = 0x00000040
	PICO_INVALID_STATE                                                           = 0x00000041
	PICO_NOT_ENOUGH_SEGMENTS                                                     = 0x00000042
	PICO_DRIVER_FUNCTION                                                         = 0x00000043
	PICO_RESERVED                                                                = 0x00000044
	PICO_INVALID_COUPLING                                                        = 0x00000045
	PICO_BUFFERS_NOT_SET                                                         = 0x00000046
	PICO_RATIO_MODE_NOT_SUPPORTED                                                = 0x00000047
	PICO_RAPID_NOT_SUPPORT_AGGREGATION                                           = 0x00000048
	PICO_INVALID_TRIGGER_PROPERTY                                                = 0x00000049
	PICO_INTERFACE_NOT_CONNECTED                                                 = 0x0000004A
	PICO_RESISTANCE_AND_PROBE_NOT_ALLOWED                                        = 0x0000004B
	PICO_POWER_FAILED                                                            = 0x0000004C
	PICO_SIGGEN_WAVEFORM_SETUP_FAILED                                            = 0x0000004D
	PICO_FPGA_FAIL                                                               = 0x0000004E
	PICO_POWER_MANAGER                                                           = 0x0000004F
	PICO_INVALID_ANALOGUE_OFFSET                                                 = 0x00000050
	PICO_PLL_LOCK_FAILED                                                         = 0x00000051
	PICO_ANALOG_BOARD                                                            = 0x00000052
	PICO_CONFIG_FAIL_AWG                                                         = 0x00000053
	PICO_INITIALISE_FPGA                                                         = 0x00000054
	PICO_EXTERNAL_FREQUENCY_INVALID                                              = 0x00000056
	PICO_CLOCK_CHANGE_ERROR                                                      = 0x00000057
	PICO_TRIGGER_AND_EXTERNAL_CLOCK_CLASH                                        = 0x00000058
	PICO_PWQ_AND_EXTERNAL_CLOCK_CLASH                                            = 0x00000059
	PICO_UNABLE_TO_OPEN_SCALING_FILE                                             = 0x0000005A
	PICO_MEMORY_CLOCK_FREQUENCY                                                  = 0x0000005B
	PICO_I2C_NOT_RESPONDING                                                      = 0x0000005C
	PICO_NO_CAPTURES_AVAILABLE                                                   = 0x0000005D
	PICO_TOO_MANY_TRIGGER_CHANNELS_IN_USE                                        = 0x0000005F
	PICO_INVALID_TRIGGER_DIRECTION                                               = 0x00000060
	PICO_INVALID_TRIGGER_STATES                                                  = 0x00000061
	PICO_NOT_USED_IN_THIS_CAPTURE_MODE                                           = 0x0000005E
	PICO_GET_DATA_ACTIVE                                                         = 0x00000103
	PICO_IP_NETWORKED                                                            = 0x00000104
	PICO_INVALID_IP_ADDRESS                                                      = 0x00000105
	PICO_IPSOCKET_FAILED                                                         = 0x00000106
	PICO_IPSOCKET_TIMEDOUT                                                       = 0x00000107
	PICO_SETTINGS_FAILED                                                         = 0x00000108
	PICO_NETWORK_FAILED                                                          = 0x00000109
	PICO_WS2_32_DLL_NOT_LOADED                                                   = 0x0000010A
	PICO_INVALID_IP_PORT                                                         = 0x0000010B
	PICO_COUPLING_NOT_SUPPORTED                                                  = 0x0000010C
	PICO_BANDWIDTH_NOT_SUPPORTED                                                 = 0x0000010D
	PICO_INVALID_BANDWIDTH                                                       = 0x0000010E
	PICO_AWG_NOT_SUPPORTED                                                       = 0x0000010F
	PICO_ETS_NOT_RUNNING                                                         = 0x00000110
	PICO_SIG_GEN_WHITENOISE_NOT_SUPPORTED                                        = 0x00000111
	PICO_SIG_GEN_WAVETYPE_NOT_SUPPORTED                                          = 0x00000112
	PICO_INVALID_DIGITAL_PORT                                                    = 0x00000113
	PICO_INVALID_DIGITAL_CHANNEL                                                 = 0x00000114
	PICO_INVALID_DIGITAL_TRIGGER_DIRECTION                                       = 0x00000115
	PICO_SIG_GEN_PRBS_NOT_SUPPORTED                                              = 0x00000116
	PICO_ETS_NOT_AVAILABLE_WITH_LOGIC_CHANNELS                                   = 0x00000117
	PICO_WARNING_REPEAT_VALUE                                                    = 0x00000118
	PICO_POWER_SUPPLY_CONNECTED                                                  = 0x00000119
	PICO_POWER_SUPPLY_NOT_CONNECTED                                              = 0x0000011A
	PICO_POWER_SUPPLY_REQUEST_INVALID                                            = 0x0000011B
	PICO_POWER_SUPPLY_UNDERVOLTAGE                                               = 0x0000011C
	PICO_CAPTURING_DATA                                                          = 0x0000011D
	PICO_USB3_0_DEVICE_NON_USB3_0_PORT                                           = 0x0000011E
	PICO_NOT_SUPPORTED_BY_THIS_DEVICE                                            = 0x0000011F
	PICO_INVALID_DEVICE_RESOLUTION                                               = 0x00000120
	PICO_INVALID_NUMBER_CHANNELS_FOR_RESOLUTION                                  = 0x00000121
	PICO_CHANNEL_DISABLED_DUE_TO_USB_POWERED                                     = 0x00000122
	PICO_SIGGEN_DC_VOLTAGE_NOT_CONFIGURABLE                                      = 0x00000123
	PICO_NO_TRIGGER_ENABLED_FOR_TRIGGER_IN_PRE_TRIG                              = 0x00000124
	PICO_TRIGGER_WITHIN_PRE_TRIG_NOT_ARMED                                       = 0x00000125
	PICO_TRIGGER_WITHIN_PRE_NOT_ALLOWED_WITH_DELAY                               = 0x00000126
	PICO_TRIGGER_INDEX_UNAVAILABLE                                               = 0x00000127
	PICO_AWG_CLOCK_FREQUENCY                                                     = 0x00000128
	PICO_TOO_MANY_CHANNELS_IN_USE                                                = 0x00000129
	PICO_NULL_CONDITIONS                                                         = 0x0000012A
	PICO_DUPLICATE_CONDITION_SOURCE                                              = 0x0000012B
	PICO_INVALID_CONDITION_INFO                                                  = 0x0000012C
	PICO_SETTINGS_READ_FAILED                                                    = 0x0000012D
	PICO_SETTINGS_WRITE_FAILED                                                   = 0x0000012E
	PICO_ARGUMENT_OUT_OF_RANGE                                                   = 0x0000012F
	PICO_HARDWARE_VERSION_NOT_SUPPORTED                                          = 0x00000130
	PICO_DIGITAL_HARDWARE_VERSION_NOT_SUPPORTED                                  = 0x00000131
	PICO_ANALOGUE_HARDWARE_VERSION_NOT_SUPPORTED                                 = 0x00000132
	PICO_UNABLE_TO_CONVERT_TO_RESISTANCE                                         = 0x00000133
	PICO_DUPLICATED_CHANNEL                                                      = 0x00000134
	PICO_INVALID_RESISTANCE_CONVERSION                                           = 0x00000135
	PICO_INVALID_VALUE_IN_MAX_BUFFER                                             = 0x00000136
	PICO_INVALID_VALUE_IN_MIN_BUFFER                                             = 0x00000137
	PICO_SIGGEN_FREQUENCY_OUT_OF_RANGE                                           = 0x00000138
	PICO_EEPROM2_CORRUPT                                                         = 0x00000139
	PICO_EEPROM2_FAIL                                                            = 0x0000013A
	PICO_SERIAL_BUFFER_TOO_SMALL                                                 = 0x0000013B
	PICO_SIGGEN_TRIGGER_AND_EXTERNAL_CLOCK_CLASH                                 = 0x0000013C
	PICO_WARNING_SIGGEN_AUXIO_TRIGGER_DISABLED                                   = 0x0000013D
	PICO_SIGGEN_GATING_AUXIO_NOT_AVAILABLE                                       = 0x00000013E
	PICO_SIGGEN_GATING_AUXIO_ENABLED                                             = 0x00000013F
	PICO_RESOURCE_ERROR                                                          = 0x00000140
	PICO_TEMPERATURE_TYPE_INVALID                                                = 0x000000141
	PICO_TEMPERATURE_TYPE_NOT_SUPPORTED                                          = 0x000000142
	PICO_TIMEOUT                                                                 = 0x00000143
	PICO_DEVICE_NOT_FUNCTIONING                                                  = 0x00000144
	PICO_INTERNAL_ERROR                                                          = 0x00000145
	PICO_MULTIPLE_DEVICES_FOUND                                                  = 0x00000146
	PICO_WARNING_NUMBER_OF_SEGMENTS_REDUCED                                      = 0x00000147
	PICO_CAL_PINS_STATES                                                         = 0x00000148
	PICO_CAL_PINS_FREQUENCY                                                      = 0x00000149
	PICO_CAL_PINS_AMPLITUDE                                                      = 0x0000014A
	PICO_CAL_PINS_WAVETYPE                                                       = 0x0000014B
	PICO_CAL_PINS_OFFSET                                                         = 0x0000014C
	PICO_PROBE_FAULT                                                             = 0x0000014D
	PICO_PROBE_IDENTITY_UNKNOWN                                                  = 0x0000014E
	PICO_PROBE_POWER_DC_POWER_SUPPLY_REQUIRED                                    = 0x0000014F
	PICO_PROBE_NOT_POWERED_WITH_DC_POWER_SUPPLY                                  = 0x00000150
	PICO_PROBE_CONFIG_FAILURE                                                    = 0x00000151
	PICO_PROBE_INTERACTION_CALLBACK                                              = 0x00000152
	PICO_UNKNOWN_INTELLIGENT_PROBE                                               = 0x00000153
	PICO_INTELLIGENT_PROBE_CORRUPT                                               = 0x00000154
	PICO_PROBE_COLLECTION_NOT_STARTED                                            = 0x00000155
	PICO_PROBE_POWER_CONSUMPTION_EXCEEDED                                        = 0x00000156
	PICO_WARNING_PROBE_CHANNEL_OUT_OF_SYNC                                       = 0x00000157
	PICO_ENDPOINT_MISSING                                                        = 0x00000158
	PICO_UNKNOWN_ENDPOINT_REQUEST                                                = 0x00000159
	PICO_ADC_TYPE_ERROR                                                          = 0x0000015A
	PICO_FPGA2_FAILED                                                            = 0x0000015B
	PICO_FPGA2_DEVICE_STATUS                                                     = 0x0000015C
	PICO_ENABLE_PROGRAM_FPGA2_FAILED                                             = 0x0000015D
	PICO_NO_CHANNELS_OR_PORTS_ENABLED                                            = 0x0000015E
	PICO_INVALID_RATIO_MODE                                                      = 0x0000015F
	PICO_READS_NOT_SUPPORTED_IN_CURRENT_CAPTURE_MODE                             = 0x00000160
	PICO_TRIGGER_READ_SELECTION_CHECK_FAILED                                     = 0x00000161
	PICO_DATA_READ1_SELECTION_CHECK_FAILED                                       = 0x00000162
	PICO_DATA_READ2_SELECTION_CHECK_FAILED                                       = 0x00000164
	PICO_DATA_READ3_SELECTION_CHECK_FAILED                                       = 0x00000168
	PICO_READ_SELECTION_OUT_OF_RANGE                                             = 0x00000170
	PICO_MULTIPLE_RATIO_MODES                                                    = 0x00000171
	PICO_NO_SAMPLES_READ                                                         = 0x00000172
	PICO_RATIO_MODE_NOT_REQUESTED                                                = 0x00000173
	PICO_NO_USER_READ_REQUESTS_SET                                               = 0x00000174
	PICO_ZERO_SAMPLES_INVALID                                                    = 0x00000175
	PICO_ANALOGUE_HARDWARE_MISSING                                               = 0x00000176
	PICO_ANALOGUE_HARDWARE_PINS                                                  = 0x00000177
	PICO_ANALOGUE_HARDWARE_SMPS_FAULT                                            = 0x00000178
	PICO_DIGITAL_ANALOGUE_HARDWARE_CONFLICT                                      = 0x00000179
	PICO_RATIO_MODE_BUFFER_NOT_SET                                               = 0x0000017A
	PICO_RESOLUTION_NOT_SUPPORTED_BY_VARIANT                                     = 0x0000017B
	PICO_THRESHOLD_OUT_OF_RANGE                                                  = 0x0000017C
	PICO_INVALID_SIMPLE_TRIGGER_DIRECTION                                        = 0x0000017D
	PICO_AUX_NOT_SUPPORTED                                                       = 0x0000017E
	PICO_NULL_DIRECTIONS                                                         = 0x0000017F
	PICO_NULL_CHANNEL_PROPERTIES                                                 = 0x00000180
	PICO_TRIGGER_CHANNEL_NOT_ENABLED                                             = 0x00000181
	PICO_CONDITION_HAS_NO_TRIGGER_PROPERTY                                       = 0x00000182
	PICO_RATIO_MODE_TRIGGER_MASKING_INVALID                                      = 0x00000183
	PICO_TRIGGER_DATA_REQUIRES_MIN_BUFFER_SIZE_OF_40_SAMPLES                     = 0x00000184
	PICO_NO_OF_CAPTURES_OUT_OF_RANGE                                             = 0x00000185
	PICO_RATIO_MODE_SEGMENT_HEADER_DOES_NOT_REQUIRE_BUFFERS                      = 0x00000186
	PICO_FOR_SEGMENT_HEADER_USE_GETTRIGGERINFO                                   = 0x00000187
	PICO_READ_NOT_SET                                                            = 0x00000188
	PICO_ADC_SETTING_MISMATCH                                                    = 0x00000189
	PICO_DATATYPE_INVALID                                                        = 0x0000018A
	PICO_RATIO_MODE_DOES_NOT_SUPPORT_DATATYPE                                    = 0x0000018B
	PICO_CHANNEL_COMBINATION_NOT_VALID_IN_THIS_RESOLUTION                        = 0x0000018C
	PICO_USE_8BIT_RESOLUTION                                                     = 0x0000018D
	PICO_AGGREGATE_BUFFERS_SAME_POINTER                                          = 0x0000018E
	PICO_OVERLAPPED_READ_VALUES_OUT_OF_RANGE                                     = 0x0000018F
	PICO_OVERLAPPED_READ_SEGMENTS_OUT_OF_RANGE                                   = 0x00000190
	PICO_CHANNELFLAGSCOMBINATIONS_ARRAY_SIZE_TOO_SMALL                           = 0x00000191
	PICO_CAPTURES_EXCEEDS_NO_OF_SUPPORTED_SEGMENTS                               = 0x00000192
	PICO_TIME_UNITS_OUT_OF_RANGE                                                 = 0x00000193
	PICO_NO_SAMPLES_REQUESTED                                                    = 0x00000194
	PICO_INVALID_ACTION                                                          = 0x00000195
	PICO_NO_OF_SAMPLES_NEED_TO_BE_EQUAL_WHEN_ADDING_BUFFERS                      = 0x00000196
	PICO_WAITING_FOR_DATA_BUFFERS                                                = 0x00000197
	PICO_STREAMING_ONLY_SUPPORTS_ONE_READ                                        = 0x00000198
	PICO_CLEAR_DATA_BUFFER_INVALID                                               = 0x00000199
	PICO_INVALID_ACTION_FLAGS_COMBINATION                                        = 0x0000019A
	PICO_BOTH_MIN_AND_MAX_NULL_BUFFERS_CANNOT_BE_ADDED                           = 0x0000019B
	PICO_CONFLICT_IN_SET_DATA_BUFFERS_CALL_REMOVE_DATA_BUFFER_TO_RESET           = 0x0000019C
	PICO_REMOVING_DATA_BUFFER_ENTRIES_NOT_ALLOWED_WHILE_DATA_PROCESSING          = 0x0000019D
	PICO_TOO_MANY_FREQUENCY_COUNTERS                                             = 0x0000019E
	PICO_CYUSB_REQUEST_FAILED                                                    = 0x00000200
	PICO_STREAMING_DATA_REQUIRED                                                 = 0x00000201
	PICO_INVALID_NUMBER_OF_SAMPLES                                               = 0x00000202
	PICO_INVALID_DISTRIBUTION                                                    = 0x00000203
	PICO_BUFFER_LENGTH_GREATER_THAN_INT32_T                                      = 0x00000204
	PICO_PLL_MUX_OUT_FAILED                                                      = 0x00000209
	PICO_ONE_PULSE_WIDTH_DIRECTION_ALLOWED                                       = 0x0000020A
	PICO_EXTERNAL_TRIGGER_NOT_SUPPORTED                                          = 0x0000020B
	PICO_NO_TRIGGER_CONDITIONS_SET                                               = 0x0000020C
	PICO_NO_OF_CHANNEL_TRIGGER_PROPERTIES_OUT_OF_RANGE                           = 0x0000020D
	PICO_PROBE_COMPONENT_ERROR                                                   = 0x0000020E
	PICO_INCOMPATIBLE_PROBE                                                      = 0x0000020F
	PICO_INVALID_TRIGGER_CHANNEL_FOR_ETS                                         = 0x00000210
	PICO_NOT_AVAILABLE_WHEN_STREAMING_IS_RUNNING                                 = 0x00000211
	PICO_INVALID_TRIGGER_WITHIN_PRE_TRIGGER_STATE                                = 0x00000212
	PICO_ZERO_NUMBER_OF_CAPTURES_INVALID                                         = 0x00000213
	PICO_INVALID_LENGTH                                                          = 0x00000214
	PICO_TRIGGER_DELAY_OUT_OF_RANGE                                              = 0x00000300
	PICO_INVALID_THRESHOLD_DIRECTION                                             = 0x00000301
	PICO_INVALID_THRESHOLD_MODE                                                  = 0x00000302
	PICO_TIMEBASE_NOT_SUPPORTED_BY_RESOLUTION                                    = 0x00000303
	PICO_INVALID_VARIANT                                                         = 0x00001000
	PICO_MEMORY_MODULE_ERROR                                                     = 0x00001001
	PICO_PULSE_WIDTH_QUALIFIER_LOWER_UPPER_CONFILCT                              = 0x00002000
	PICO_PULSE_WIDTH_QUALIFIER_TYPE                                              = 0x00002001
	PICO_PULSE_WIDTH_QUALIFIER_DIRECTION                                         = 0x00002002
	PICO_THRESHOLD_MODE_OUT_OF_RANGE                                             = 0x00002003
	PICO_TRIGGER_AND_PULSEWIDTH_DIRECTION_IN_CONFLICT                            = 0x00002004
	PICO_THRESHOLD_UPPER_LOWER_MISMATCH                                          = 0x00002005
	PICO_PULSE_WIDTH_LOWER_OUT_OF_RANGE                                          = 0x00002006
	PICO_PULSE_WIDTH_UPPER_OUT_OF_RANGE                                          = 0x00002007
	PICO_FRONT_PANEL_ERROR                                                       = 0x00002008
	PICO_FRONT_PANEL_MODE                                                        = 0x0000200B
	PICO_FRONT_PANEL_FEATURE                                                     = 0x0000200C
	PICO_NO_PULSE_WIDTH_CONDITIONS_SET                                           = 0x0000200D
	PICO_TRIGGER_PORT_NOT_ENABLED                                                = 0x0000200E
	PICO_DIGITAL_DIRECTION_NOT_SET                                               = 0x0000200F
	PICO_I2C_DEVICE_INVALID_READ_COMMAND                                         = 0x00002010
	PICO_I2C_DEVICE_INVALID_RESPONSE                                             = 0x00002011
	PICO_I2C_DEVICE_INVALID_WRITE_COMMAND                                        = 0x00002012
	PICO_I2C_DEVICE_ARGUMENT_OUT_OF_RANGE                                        = 0x00002013
	PICO_I2C_DEVICE_MODE                                                         = 0x00002014
	PICO_I2C_DEVICE_SETUP_FAILED                                                 = 0x00002015
	PICO_I2C_DEVICE_FEATURE                                                      = 0x00002016
	PICO_I2C_DEVICE_VALIDATION_FAILED                                            = 0x00002017
	PICO_INTERNAL_HEADER_ERROR                                                   = 0x00002018
	PICO_FAILED_TO_WRITE_HARDWARE_FAULT                                          = 0x00002019
	PICO_MSO_TOO_MANY_EDGE_TRANSITIONS_WHEN_USING_PULSE_WIDTH                    = 0x00003000
	PICO_INVALID_PROBE_LED_POSITION                                              = 0x00003001
	PICO_PROBE_LED_POSITION_NOT_SUPPORTED                                        = 0x00003002
	PICO_DUPLICATE_PROBE_CHANNEL_LED_POSITION                                    = 0x00003003
	PICO_PROBE_LED_FAILURE                                                       = 0x00003004
	PICO_PROBE_NOT_SUPPORTED_BY_THIS_DEVICE                                      = 0x00003005
	PICO_INVALID_PROBE_NAME                                                      = 0x00003006
	PICO_NO_PROBE_COLOUR_SETTINGS                                                = 0x00003007
	PICO_NO_PROBE_CONNECTED_ON_REQUESTED_CHANNEL                                 = 0x00003008
	PICO_PROBE_DOES_NOT_REQUIRE_CALIBRATION                                      = 0x00003009
	PICO_PROBE_CALIBRATION_FAILED                                                = 0x0000300A
	PICO_PROBE_VERSION_ERROR                                                     = 0x0000300B
	PICO_PROBE_DOES_NOT_SUPPORT_FREQUENCY_COUNTER                                = 0x0000300C
	PICO_AUTO_TRIGGER_TIME_TOO_LONG                                              = 0x00004000
	PICO_MSO_POD_VALIDATION_FAILED                                               = 0x00005000
	PICO_NO_MSO_POD_CONNECTED                                                    = 0x00005001
	PICO_DIGITAL_PORT_HYSTERESIS_OUT_OF_RANGE                                    = 0x00005002
	PICO_MSO_POD_FAILED_UNIT                                                     = 0x00005003
	PICO_ATTENUATION_FAILED                                                      = 0x00005004
	PICO_DC_50OHM_OVERVOLTAGE_TRIPPED                                            = 0x00005005
	PICO_MSO_OVER_CURRENT_TRIPPED                                                = 0x00005006
	PICO_NOT_RESPONDING_OVERHEATED                                               = 0x00005010
	PICO_USB_VERSION_NOT_SUPPORTED                                               = 0x00005100
	PICO_HARDWARE_CAPTURE_TIMEOUT                                                = 0x00006000
	PICO_HARDWARE_READY_TIMEOUT                                                  = 0x00006001
	PICO_HARDWARE_CAPTURING_CALL_STOP                                            = 0x00006002
	PICO_TOO_FEW_REQUESTED_STREAMING_SAMPLES                                     = 0x00007000
	PICO_STREAMING_REREAD_DATA_NOT_AVAILABLE                                     = 0x00007001
	PICO_STREAMING_COMBINATION_OF_RAW_DATA_AND_ONE_AGGREGATION_DATA_TYPE_ALLOWED = 0x00007002
	PICO_DEVICE_TIME_STAMP_RESET                                                 = 0x01000000
	PICO_TRIGGER_TIME_NOT_REQUESTED                                              = 0x02000001
	PICO_TRIGGER_TIME_BUFFER_NOT_SET                                             = 0x02000002
	PICO_TRIGGER_TIME_FAILED_TO_CALCULATE                                        = 0x02000004
	PICO_TRIGGER_WITHIN_A_PRE_TRIGGER_FAILED_TO_CALCULATE                        = 0x02000008
	PICO_TRIGGER_TIME_STAMP_NOT_REQUESTED                                        = 0x02000100
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_REQUIRE_BUFFERS   = 0x02200000
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_HAVE_BUFFERS      = 0x02200001
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_USE_GETTRIGGERINFO         = 0x02200002
	PICO_STREAMING_DOES_NOT_SUPPORT_TRIGGER_RATIO_MODES                          = 0x02200003
	PICO_USE_THE_TRIGGER_READ                                                    = 0x02200004
	PICO_USE_A_DATA_READ                                                         = 0x02200005
	PICO_TRIGGER_READ_REQUIRES_INT16_T_DATA_TYPE                                 = 0x02200006
	PICO_RATIO_MODE_REQUIRES_NUMBER_OF_SAMPLES_TO_BE_SET                         = 0x02200007
	PICO_SIGGEN_SETTINGS_MISMATCH                                                = 0x03000010
	PICO_SIGGEN_SETTINGS_CHANGED_CALL_APPLY                                      = 0x03000011
	PICO_SIGGEN_WAVETYPE_NOT_SUPPORTED                                           = 0x03000012
	PICO_SIGGEN_TRIGGERTYPE_NOT_SUPPORTED                                        = 0x03000013
	PICO_SIGGEN_TRIGGERSOURCE_NOT_SUPPORTED                                      = 0x03000014
	PICO_SIGGEN_FILTER_STATE_NOT_SUPPORTED                                       = 0x03000015
	PICO_SIGGEN_NULL_PARAMETER                                                   = 0x03000020
	PICO_SIGGEN_EMPTY_BUFFER_SUPPLIED                                            = 0x03000021
	PICO_SIGGEN_RANGE_NOT_SUPPLIED                                               = 0x03000022
	PICO_SIGGEN_BUFFER_NOT_SUPPLIED                                              = 0x03000023
	PICO_SIGGEN_FREQUENCY_NOT_SUPPLIED                                           = 0x03000024
	PICO_SIGGEN_SWEEP_INFO_NOT_SUPPLIED                                          = 0x03000025
	PICO_SIGGEN_TRIGGER_INFO_NOT_SUPPLIED                                        = 0x03000026
	PICO_SIGGEN_CLOCK_FREQ_NOT_SUPPLIED                                          = 0x03000027
	PICO_SIGGEN_TOO_MANY_SAMPLES                                                 = 0x03000030
	PICO_SIGGEN_DUTYCYCLE_OUT_OF_RANGE                                           = 0x03000031
	PICO_SIGGEN_CYCLES_OUT_OF_RANGE                                              = 0x03000032
	PICO_SIGGEN_PRESCALE_OUT_OF_RANGE                                            = 0x03000033
	PICO_SIGGEN_SWEEPTYPE_INVALID                                                = 0x03000034
	PICO_SIGGEN_SWEEP_WAVETYPE_MISMATCH                                          = 0x03000035
	PICO_SIGGEN_INVALID_SWEEP_PARAMETERS                                         = 0x03000036
	PICO_SIGGEN_SWEEP_PRESCALE_NOT_SUPPORTED                                     = 0x03000037
	PICO_AWG_OVER_VOLTAGE_RANGE                                                  = 0x03000038
	PICO_NOT_LOCKED_TO_REFERENCE_FREQUENCY                                       = 0x03000039
	PICO_PERMISSIONS_ERROR                                                       = 0x03000040
	PICO_PORTS_WITHOUT_ANALOGUE_CHANNELS_ONLY_ALLOWED_IN_8BIT_RESOLUTION         = 0x03001000
	PICO_ANALOGUE_FRONTEND_MISSING                                               = 0x03003001
	PICO_FRONT_PANEL_MISSING                                                     = 0x03003002
	PICO_ANALOGUE_FRONTEND_AND_FRONT_PANEL_MISSING                               = 0x03003003
	PICO_DIGITAL_BOARD_HARDWARE_ERROR                                            = 0x03003800
	PICO_FIRMWARE_UPDATE_REQUIRED_TO_USE_DEVICE_WITH_THIS_DRIVER                 = 0x03004000
	PICO_UPDATE_REQUIRED_NULL                                                    = 0x03004001
	PICO_FIRMWARE_UP_TO_DATE                                                     = 0x03004002
	PICO_FLASH_FAIL                                                              = 0x03004003
	PICO_INTERNAL_ERROR_FIRMWARE_LENGTH_INVALID                                  = 0x03004004
	PICO_INTERNAL_ERROR_FIRMWARE_NULL                                            = 0x03004005
	PICO_FIRMWARE_FAILED_TO_BE_CHANGED                                           = 0x03004006
	PICO_FIRMWARE_FAILED_TO_RELOAD                                               = 0x03004007
	PICO_FIRMWARE_FAILED_TO_BE_UPDATE                                            = 0x03004008
	PICO_FIRMWARE_VERSION_OUT_OF_RANGE                                           = 0x03004009
	PICO_OPTIONAL_BOOTLOADER_UPDATE_AVAILABLE_WITH_THIS_DRIVER                   = 0x03005000
	PICO_BOOTLOADER_VERSION_NOT_AVAILABLE                                        = 0x03005001
	PICO_NO_APPS_AVAILABLE                                                       = 0x03008000
	PICO_UNSUPPORTED_APP                                                         = 0x03008001
	PICO_ADC_POWERED_DOWN                                                        = 0x03002000
	PICO_WATCHDOGTIMER                                                           = 0x10000000
	PICO_IPP_NOT_FOUND                                                           = 0x10000001
	PICO_IPP_NO_FUNCTION                                                         = 0x10000002
	PICO_IPP_ERROR                                                               = 0x10000003
	PICO_SHADOW_CAL_NOT_AVAILABLE                                                = 0x10000004
	PICO_SHADOW_CAL_DISABLED                                                     = 0x10000005
	PICO_SHADOW_CAL_ERROR                                                        = 0x10000006
	PICO_SHADOW_CAL_CORRUPT                                                      = 0x10000007
	PICO_DEVICE_MEMORY_OVERFLOW                                                  = 0x10000008
	PICO_ADC_TEST_FAILURE                                                        = 0x10000010
	PICO_RESERVED_1                                                              = 0x11000000
	PICO_SOURCE_NOT_READY                                                        = 0x20000000
	PICO_SOURCE_INVALID_BAUD_RATE                                                = 0x20000001
	PICO_SOURCE_NOT_OPENED_FOR_WRITE                                             = 0x20000002
	PICO_SOURCE_FAILED_TO_WRITE_DEVICE                                           = 0x20000003
	PICO_SOURCE_EEPROM_FAIL                                                      = 0x20000004
	PICO_SOURCE_EEPROM_NOT_PRESENT                                               = 0x20000005
	PICO_SOURCE_EEPROM_NOT_PROGRAMMED                                            = 0x20000006
	PICO_SOURCE_LIST_NOT_READY                                                   = 0x20000007
	PICO_SOURCE_FTD2XX_NOT_FOUND                                                 = 0x20000008
	PICO_SOURCE_FTD2XX_NO_FUNCTION                                               = 0x20000009

	PicoDriverVersion             = PICO_DRIVER_VERSION
	PicoUsbVersion                = PICO_USB_VERSION
	PicoHardwareVersion           = PICO_HARDWARE_VERSION
	PicoVariantIfo                = PICO_VARIANT_INFO
	PicoBatchAndSerial            = PICO_BATCH_AND_SERIAL
	PicoCalDate                   = PICO_CAL_DATE
	PicoKernelVersion             = PICO_KERNEL_VERSION
	PicoDigitalHardwareVersion    = PICO_DIGITAL_HARDWARE_VERSION
	PicoAnaloguelHardwareVersion  = PICO_ANALOGUE_HARDWARE_VERSION
	PicoFirmwareVersion1          = PICO_FIRMWARE_VERSION_1
	PicoFirmwareVersion2          = PICO_FIRMWARE_VERSION_2
	PicoFirmwareVersion3          = PICO_FIRMWARE_VERSION_3
	PicoMacAddress                = PICO_MAC_ADDRESS
	PicoShadowCal                 = PICO_SHADOW_CAL
	PicoIppVersion                = PICO_IPP_VERSION
	PicoDriverPath                = PICO_DRIVER_PATH
	PicoFrontPanelFirmwareVersion = PICO_FRONT_PANEL_FIRMWARE_VERSION
	PicoBootloaderVersion         = PICO_BOOTLOADER_VERSION
	PicoOk                        = PICO_OK
)

var statMap = map[int]string{
	PICO_OK:                                                  "PICO_OK  The PicoScope is functioning correctly.",
	PICO_MAX_UNITS_OPENED:                                    "PICO_MAX_UNITS_OPENED  An attempt has been made to open more than <API>_MAX_UNITS.",
	PICO_MEMORY_FAIL:                                         "PICO_MEMORY_FAIL  Not enough memory could be allocated on the host machine.",
	PICO_NOT_FOUND:                                           "PICO_NOT_FOUND  No Pico Technology device could be found.",
	PICO_FW_FAIL:                                             "PICO_FW_FAIL  Unable to download firmware.",
	PICO_OPEN_OPERATION_IN_PROGRESS:                          "PICO_OPEN_OPERATION_IN_PROGRESS  The driver is busy opening a device.",
	PICO_OPERATION_FAILED:                                    "PICO_OPERATION_FAILED  An unspecified failure occurred.",
	PICO_NOT_RESPONDING:                                      "PICO_NOT_RESPONDING  The PicoScope is not responding to commands from the PC.",
	PICO_CONFIG_FAIL:                                         "PICO_CONFIG_FAIL  The configuration information in the PicoScope is corrupt or missing.",
	PICO_KERNEL_DRIVER_TOO_OLD:                               "PICO_KERNEL_DRIVER_TOO_OLD  The picopp.sys file is too old to be used with the device driver.",
	PICO_EEPROM_CORRUPT:                                      "PICO_EEPROM_CORRUPT  The EEPROM has become corrupt, so the device will use a default setting.",
	PICO_OS_NOT_SUPPORTED:                                    "PICO_OS_NOT_SUPPORTED  The operating system on the PC is not supported by this driver.",
	PICO_INVALID_HANDLE:                                      "PICO_INVALID_HANDLE  There is no device with the handle value passed.",
	PICO_INVALID_PARAMETER:                                   "PICO_INVALID_PARAMETER  A parameter value is not valid.",
	PICO_INVALID_TIMEBASE:                                    "PICO_INVALID_TIMEBASE  The timebase is not supported or is invalid.",
	PICO_INVALID_VOLTAGE_RANGE:                               "PICO_INVALID_VOLTAGE_RANGE  The voltage range is not supported or is invalid.",
	PICO_INVALID_CHANNEL:                                     "PICO_INVALID_CHANNEL  The channel number is not valid on this device or no channels have been set.",
	PICO_INVALID_TRIGGER_CHANNEL:                             "PICO_INVALID_TRIGGER_CHANNEL  The channel set for a trigger is not available on this device.",
	PICO_INVALID_CONDITION_CHANNEL:                           "PICO_INVALID_CONDITION_CHANNEL  The channel set for a condition is not available on this device.",
	PICO_NO_SIGNAL_GENERATOR:                                 "PICO_NO_SIGNAL_GENERATOR  The device does not have a signal generator.",
	PICO_STREAMING_FAILED:                                    "PICO_STREAMING_FAILED  Streaming has failed to start or has stopped without user request.",
	PICO_BLOCK_MODE_FAILED:                                   "PICO_BLOCK_MODE_FAILED  Block failed to start - a parameter may have been set wrongly.",
	PICO_NULL_PARAMETER:                                      "PICO_NULL_PARAMETER  A parameter that was required is NULL.",
	PICO_ETS_MODE_SET:                                        "PICO_ETS_MODE_SET  The current functionality is not available while using ETS capture mode.",
	PICO_DATA_NOT_AVAILABLE:                                  "PICO_DATA_NOT_AVAILABLE  No data is available from a run block call.",
	PICO_STRING_BUFFER_TO_SMALL:                              "PICO_STRING_BUFFER_TO_SMALL  The buffer passed for the information was too small.",
	PICO_ETS_NOT_SUPPORTED:                                   "PICO_ETS_NOT_SUPPORTED  ETS is not supported on this device.",
	PICO_AUTO_TRIGGER_TIME_TO_SHORT:                          "PICO_AUTO_TRIGGER_TIME_TO_SHORT  The auto trigger time is less than the time it will take to collect the pre-trigger data.",
	PICO_BUFFER_STALL:                                        "PICO_BUFFER_STALL  The collection of data has stalled as unread data would be overwritten.",
	PICO_TOO_MANY_SAMPLES:                                    "PICO_TOO_MANY_SAMPLES  Number of samples requested is more than available in the current memory segment.",
	PICO_TOO_MANY_SEGMENTS:                                   "PICO_TOO_MANY_SEGMENTS  Not possible to create number of segments requested.",
	PICO_PULSE_WIDTH_QUALIFIER:                               "PICO_PULSE_WIDTH_QUALIFIER  A null pointer has been passed in the trigger function or one of the parameters is out of range.",
	PICO_DELAY:                                               "PICO_DELAY  One or more of the hold-off parameters are out of range.",
	PICO_SOURCE_DETAILS:                                      "PICO_SOURCE_DETAILS  One or more of the source details are incorrect.",
	PICO_CONDITIONS:                                          "PICO_CONDITIONS  One or more of the conditions are incorrect.",
	PICO_USER_CALLBACK:                                       "PICO_USER_CALLBACK  The driver's thread is currently in the <API>Ready callback function and therefore the action cannot be carried out.",
	PICO_DEVICE_SAMPLING:                                     "PICO_DEVICE_SAMPLING  An attempt is being made to get stored data while streaming. Either stop streaming by calling <API>Stop, or use <API>GetStreamingLatestValues.",
	PICO_NO_SAMPLES_AVAILABLE:                                "PICO_NO_SAMPLES_AVAILABLE  Data is unavailable because a run has not been completed.",
	PICO_SEGMENT_OUT_OF_RANGE:                                "PICO_SEGMENT_OUT_OF_RANGE  The memory segment index is out of range.",
	PICO_BUSY:                                                "PICO_BUSY  The device is busy so data cannot be returned yet.",
	PICO_STARTINDEX_INVALID:                                  "PICO_STARTINDEX_INVALID  The start time to get stored data is out of range.",
	PICO_INVALID_INFO:                                        "PICO_INVALID_INFO  The information number requested is not a valid number.",
	PICO_INFO_UNAVAILABLE:                                    "PICO_INFO_UNAVAILABLE  The handle is invalid so no information is available about the device. Only PICO_DRIVER_VERSION is available.",
	PICO_INVALID_SAMPLE_INTERVAL:                             "PICO_INVALID_SAMPLE_INTERVAL  The sample interval selected for streaming is out of range.",
	PICO_TRIGGER_ERROR:                                       "PICO_TRIGGER_ERROR  ETS is set but no trigger has been set. A trigger setting is required for ETS.",
	PICO_MEMORY:                                              "PICO_MEMORY  Driver cannot allocate memory.",
	PICO_SIG_GEN_PARAM:                                       "PICO_SIG_GEN_PARAM  Incorrect parameter passed to the signal generator.",
	PICO_SHOTS_SWEEPS_WARNING:                                "PICO_SHOTS_SWEEPS_WARNING  Conflict between the shots and sweeps parameters sent to the signal generator.",
	PICO_SIGGEN_TRIGGER_SOURCE:                               "PICO_SIGGEN_TRIGGER_SOURCE  A software trigger has been sent but the trigger source is not a software trigger.",
	PICO_AUX_OUTPUT_CONFLICT:                                 "PICO_AUX_OUTPUT_CONFLICT  An <API>SetTrigger call has found a conflict between the trigger source and the AUX output enable.",
	PICO_AUX_OUTPUT_ETS_CONFLICT:                             "PICO_AUX_OUTPUT_ETS_CONFLICT  ETS mode is being used and AUX is set as an input.",
	PICO_WARNING_EXT_THRESHOLD_CONFLICT:                      "PICO_WARNING_EXT_THRESHOLD_CONFLICT  Attempt to set different EXT input thresholds set for signal generator and oscilloscope trigger.",
	PICO_WARNING_AUX_OUTPUT_CONFLICT:                         "PICO_WARNING_AUX_OUTPUT_CONFLICT  An <API>SetTrigger... function has set AUX as an output and the signal generator is using it as a trigger.",
	PICO_SIGGEN_OUTPUT_OVER_VOLTAGE:                          "PICO_SIGGEN_OUTPUT_OVER_VOLTAGE  The combined peak-to-peak voltage and the analog offset voltage exceed the maximum voltage the signal generator can produce.",
	PICO_DELAY_NULL:                                          "PICO_DELAY_NULL  NULL pointer passed as delay parameter.",
	PICO_INVALID_BUFFER:                                      "PICO_INVALID_BUFFER  The buffers for overview data have not been set while streaming.",
	PICO_SIGGEN_OFFSET_VOLTAGE:                               "PICO_SIGGEN_OFFSET_VOLTAGE  The analog offset voltage is out of range.",
	PICO_SIGGEN_PK_TO_PK:                                     "PICO_SIGGEN_PK_TO_PK  The analog peak-to-peak voltage is out of range.",
	PICO_CANCELLED:                                           "PICO_CANCELLED  A block collection has been cancelled.",
	PICO_SEGMENT_NOT_USED:                                    "PICO_SEGMENT_NOT_USED  The segment index is not currently being used.",
	PICO_INVALID_CALL:                                        "PICO_INVALID_CALL  The wrong GetValues function has been called for the collection mode in use.",
	PICO_GET_VALUES_INTERRUPTED:                              "PICO_GET_VALUES_INTERRUPTED",
	PICO_NOT_USED:                                            "PICO_NOT_USED  The function is not available.",
	PICO_INVALID_SAMPLERATIO:                                 "PICO_INVALID_SAMPLERATIO  The aggregation ratio requested is out of range.",
	PICO_INVALID_STATE:                                       "PICO_INVALID_STATE  Device is in an invalid state.",
	PICO_NOT_ENOUGH_SEGMENTS:                                 "PICO_NOT_ENOUGH_SEGMENTS  The number of segments allocated is fewer than the number of captures requested.",
	PICO_DRIVER_FUNCTION:                                     "PICO_DRIVER_FUNCTION  A driver function has already been called and not yet finished. Only one call to the driver can be made at any one time.",
	PICO_RESERVED:                                            "PICO_RESERVED  Not used.",
	PICO_INVALID_COUPLING:                                    "PICO_INVALID_COUPLING  An invalid coupling type was specified in <API>SetChannel.",
	PICO_BUFFERS_NOT_SET:                                     "PICO_BUFFERS_NOT_SET  An attempt was made to get data before a data buffer was defined.",
	PICO_RATIO_MODE_NOT_SUPPORTED:                            "PICO_RATIO_MODE_NOT_SUPPORTED  The selected downsampling mode (used for data reduction) is not allowed.",
	PICO_RAPID_NOT_SUPPORT_AGGREGATION:                       "PICO_RAPID_NOT_SUPPORT_AGGREGATION  Aggregation was requested in rapid block mode.",
	PICO_INVALID_TRIGGER_PROPERTY:                            "PICO_INVALID_TRIGGER_PROPERTY  An invalid parameter was passed to <API>SetTriggerChannelProperties(V2).",
	PICO_INTERFACE_NOT_CONNECTED:                             "PICO_INTERFACE_NOT_CONNECTED  The driver was unable to contact the oscilloscope.",
	PICO_RESISTANCE_AND_PROBE_NOT_ALLOWED:                    "PICO_RESISTANCE_AND_PROBE_NOT_ALLOWED  Resistance-measuring mode is not allowed in conjunction with the specified probe.",
	PICO_POWER_FAILED:                                        "PICO_POWER_FAILED  The device was unexpectedly powered down.",
	PICO_SIGGEN_WAVEFORM_SETUP_FAILED:                        "PICO_SIGGEN_WAVEFORM_SETUP_FAILED  A problem occurred in <API>SetSigGenBuiltIn or <API>SetSigGenArbitrary.",
	PICO_FPGA_FAIL:                                           "PICO_FPGA_FAIL  FPGA not successfully set up.",
	PICO_POWER_MANAGER:                                       "PICO_POWER_MANAGER",
	PICO_INVALID_ANALOGUE_OFFSET:                             "PICO_INVALID_ANALOGUE_OFFSET  An impossible analog offset value was specified in <API>SetChannel.",
	PICO_PLL_LOCK_FAILED:                                     "PICO_PLL_LOCK_FAILED  There is an error within the device hardware.",
	PICO_ANALOG_BOARD:                                        "PICO_ANALOG_BOARD  There is an error within the device hardware.",
	PICO_CONFIG_FAIL_AWG:                                     "PICO_CONFIG_FAIL_AWG  Unable to configure the signal generator.",
	PICO_INITIALISE_FPGA:                                     "PICO_INITIALISE_FPGA  The FPGA cannot be initialized, so unit cannot be opened.",
	PICO_EXTERNAL_FREQUENCY_INVALID:                          "PICO_EXTERNAL_FREQUENCY_INVALID  The frequency for the external clock is not within 15% of the nominal value.",
	PICO_CLOCK_CHANGE_ERROR:                                  "PICO_CLOCK_CHANGE_ERROR  The FPGA could not lock the clock signal.",
	PICO_TRIGGER_AND_EXTERNAL_CLOCK_CLASH:                    "PICO_TRIGGER_AND_EXTERNAL_CLOCK_CLASH  You are trying to configure the AUX input as both a trigger and a reference clock.",
	PICO_PWQ_AND_EXTERNAL_CLOCK_CLASH:                        "PICO_PWQ_AND_EXTERNAL_CLOCK_CLASH  You are trying to configure the AUX input as both a pulse width qualifier and a reference clock.",
	PICO_UNABLE_TO_OPEN_SCALING_FILE:                         "PICO_UNABLE_TO_OPEN_SCALING_FILE  The requested scaling file cannot be opened.",
	PICO_MEMORY_CLOCK_FREQUENCY:                              "PICO_MEMORY_CLOCK_FREQUENCY  The frequency of the memory is reporting incorrectly.",
	PICO_I2C_NOT_RESPONDING:                                  "PICO_I2C_NOT_RESPONDING  The I2C that is being actioned is not responding to requests.",
	PICO_NO_CAPTURES_AVAILABLE:                               "PICO_NO_CAPTURES_AVAILABLE  There are no captures available and therefore no data can be returned.",
	PICO_TOO_MANY_TRIGGER_CHANNELS_IN_USE:                    "PICO_TOO_MANY_TRIGGER_CHANNELS_IN_USE  The number of trigger channels is greater than 4, except for a PicoScope 4824 where 8 channels are allowed for rising/falling/rising_or_falling trigger directions.",
	PICO_INVALID_TRIGGER_DIRECTION:                           "PICO_INVALID_TRIGGER_DIRECTION  If you have specified a trigger direction which is not allowed, for example specifying PICO_ABOVE without another condition which crosses a threshold on another channel.",
	PICO_INVALID_TRIGGER_STATES:                              "PICO_INVALID_TRIGGER_STATES  When more than 4 trigger channels are set and their trigger condition states are not <API>_CONDITION_TRUE.",
	PICO_NOT_USED_IN_THIS_CAPTURE_MODE:                       "PICO_NOT_USED_IN_THIS_CAPTURE_MODE  The capture mode the device is currently running in does not support the current request.",
	PICO_GET_DATA_ACTIVE:                                     "PICO_GET_DATA_ACTIVE",
	PICO_IP_NETWORKED:                                        "PICO_IP_NETWORKED  The device is currently connected via the IP Network socket and thus the call made is not supported.",
	PICO_INVALID_IP_ADDRESS:                                  "PICO_INVALID_IP_ADDRESS  An incorrect IP address has been passed to the driver.",
	PICO_IPSOCKET_FAILED:                                     "PICO_IPSOCKET_FAILED  The IP socket has failed.",
	PICO_IPSOCKET_TIMEDOUT:                                   "PICO_IPSOCKET_TIMEDOUT  The IP socket has timed out.",
	PICO_SETTINGS_FAILED:                                     "PICO_SETTINGS_FAILED  Failed to apply the requested settings.",
	PICO_NETWORK_FAILED:                                      "PICO_NETWORK_FAILED  The network connection has failed.",
	PICO_WS2_32_DLL_NOT_LOADED:                               "PICO_WS2_32_DLL_NOT_LOADED  Unable to load the WS2 DLL.",
	PICO_INVALID_IP_PORT:                                     "PICO_INVALID_IP_PORT  The specified IP port is invalid.",
	PICO_COUPLING_NOT_SUPPORTED:                              "PICO_COUPLING_NOT_SUPPORTED  The type of coupling requested is not supported on the opened device.",
	PICO_BANDWIDTH_NOT_SUPPORTED:                             "PICO_BANDWIDTH_NOT_SUPPORTED  Bandwidth limiting is not supported on the opened device.",
	PICO_INVALID_BANDWIDTH:                                   "PICO_INVALID_BANDWIDTH  The value requested for the bandwidth limit is out of range.",
	PICO_AWG_NOT_SUPPORTED:                                   "PICO_AWG_NOT_SUPPORTED  The arbitrary waveform generator is not supported by the opened device.",
	PICO_ETS_NOT_RUNNING:                                     "PICO_ETS_NOT_RUNNING  Data has been requested with ETS mode set but run block has not been called, or stop has been called.",
	PICO_SIG_GEN_WHITENOISE_NOT_SUPPORTED:                    "PICO_SIG_GEN_WHITENOISE_NOT_SUPPORTED  White noise output is not supported on the opened device.",
	PICO_SIG_GEN_WAVETYPE_NOT_SUPPORTED:                      "PICO_SIG_GEN_WAVETYPE_NOT_SUPPORTED  The wave type requested is not supported by the opened device.",
	PICO_INVALID_DIGITAL_PORT:                                "PICO_INVALID_DIGITAL_PORT  The requested digital port number is out of range (MSOs only).",
	PICO_INVALID_DIGITAL_CHANNEL:                             "PICO_INVALID_DIGITAL_CHANNEL  The digital channel is not in the range <API>_DIGITAL_CHANNEL0 to <API>_DIGITAL_CHANNEL15, the digital channels that are supported.",
	PICO_INVALID_DIGITAL_TRIGGER_DIRECTION:                   "PICO_INVALID_DIGITAL_TRIGGER_DIRECTION  The digital trigger direction is not a valid trigger direction and should be equal in value to one of the <API>_DIGITAL_DIRECTION enumerations.",
	PICO_SIG_GEN_PRBS_NOT_SUPPORTED:                          "PICO_SIG_GEN_PRBS_NOT_SUPPORTED  Signal generator does not generate pseudo-random binary sequence.",
	PICO_ETS_NOT_AVAILABLE_WITH_LOGIC_CHANNELS:               "PICO_ETS_NOT_AVAILABLE_WITH_LOGIC_CHANNELS  When a digital port is enabled, ETS sample mode is not available for use.",
	PICO_WARNING_REPEAT_VALUE:                                "PICO_WARNING_REPEAT_VALUE  There has been no new sample taken, this value has already been returned previously.",
	PICO_POWER_SUPPLY_CONNECTED:                              "PICO_POWER_SUPPLY_CONNECTED  The DC power supply is connected.",
	PICO_POWER_SUPPLY_NOT_CONNECTED:                          "PICO_POWER_SUPPLY_NOT_CONNECTED  The DC power supply is not connected. For many 4+ Channel devices this will mean a restricted feature set is offered e.g. for a 4 channel device - C and D are usually disabled. Check the respective API programmers guide of your device for the full details.",
	PICO_POWER_SUPPLY_REQUEST_INVALID:                        "PICO_POWER_SUPPLY_REQUEST_INVALID  Incorrect power mode passed for current power source.",
	PICO_POWER_SUPPLY_UNDERVOLTAGE:                           "PICO_POWER_SUPPLY_UNDERVOLTAGE  The supply voltage from the USB source is too low.",
	PICO_CAPTURING_DATA:                                      "PICO_CAPTURING_DATA  The oscilloscope is in the process of capturing data.",
	PICO_USB3_0_DEVICE_NON_USB3_0_PORT:                       "PICO_USB3_0_DEVICE_NON_USB3_0_PORT  A USB 3.0 device is connected to a non-USB 3.0 port.",
	PICO_NOT_SUPPORTED_BY_THIS_DEVICE:                        "PICO_NOT_SUPPORTED_BY_THIS_DEVICE  A function has been called that is not supported by the current device.",
	PICO_INVALID_DEVICE_RESOLUTION:                           "PICO_INVALID_DEVICE_RESOLUTION  The device resolution is invalid (out of range).",
	PICO_INVALID_NUMBER_CHANNELS_FOR_RESOLUTION:              "PICO_INVALID_NUMBER_CHANNELS_FOR_RESOLUTION  The number of channels that can be enabled is limited in 15 and 16-bit modes. (Flexible Resolution Oscilloscopes only)",
	PICO_CHANNEL_DISABLED_DUE_TO_USB_POWERED:                 "PICO_CHANNEL_DISABLED_DUE_TO_USB_POWERED  USB power not sufficient for all requested channels.",
	PICO_SIGGEN_DC_VOLTAGE_NOT_CONFIGURABLE:                  "PICO_SIGGEN_DC_VOLTAGE_NOT_CONFIGURABLE  The signal generator does not have a configurable DC offset.",
	PICO_NO_TRIGGER_ENABLED_FOR_TRIGGER_IN_PRE_TRIG:          "PICO_NO_TRIGGER_ENABLED_FOR_TRIGGER_IN_PRE_TRIG  An attempt has been made to define pre-trigger delay without first enabling a trigger.",
	PICO_TRIGGER_WITHIN_PRE_TRIG_NOT_ARMED:                   "PICO_TRIGGER_WITHIN_PRE_TRIG_NOT_ARMED  An attempt has been made to define pre-trigger delay without first arming a trigger.",
	PICO_TRIGGER_WITHIN_PRE_NOT_ALLOWED_WITH_DELAY:           "PICO_TRIGGER_WITHIN_PRE_NOT_ALLOWED_WITH_DELAY  Pre-trigger delay and post-trigger delay cannot be used at the same time.",
	PICO_TRIGGER_INDEX_UNAVAILABLE:                           "PICO_TRIGGER_INDEX_UNAVAILABLE  The array index points to a nonexistent trigger.",
	PICO_AWG_CLOCK_FREQUENCY:                                 "PICO_AWG_CLOCK_FREQUENCY",
	PICO_TOO_MANY_CHANNELS_IN_USE:                            "PICO_TOO_MANY_CHANNELS_IN_USE  There are more than 4 analog channels with a trigger condition set.",
	PICO_NULL_CONDITIONS:                                     "PICO_NULL_CONDITIONS  The condition parameter is a null pointer.",
	PICO_DUPLICATE_CONDITION_SOURCE:                          "PICO_DUPLICATE_CONDITION_SOURCE  There is more than one condition pertaining to the same channel.",
	PICO_INVALID_CONDITION_INFO:                              "PICO_INVALID_CONDITION_INFO  The parameter relating to condition information is out of range.",
	PICO_SETTINGS_READ_FAILED:                                "PICO_SETTINGS_READ_FAILED  Reading the meta data has failed.",
	PICO_SETTINGS_WRITE_FAILED:                               "PICO_SETTINGS_WRITE_FAILED  Writing the meta data has failed.",
	PICO_ARGUMENT_OUT_OF_RANGE:                               "PICO_ARGUMENT_OUT_OF_RANGE  A parameter has a value out of the expected range.",
	PICO_HARDWARE_VERSION_NOT_SUPPORTED:                      "PICO_HARDWARE_VERSION_NOT_SUPPORTED  The driver does not support the hardware variant connected.",
	PICO_DIGITAL_HARDWARE_VERSION_NOT_SUPPORTED:              "PICO_DIGITAL_HARDWARE_VERSION_NOT_SUPPORTED  The driver does not support the digital hardware variant connected.",
	PICO_ANALOGUE_HARDWARE_VERSION_NOT_SUPPORTED:             "PICO_ANALOGUE_HARDWARE_VERSION_NOT_SUPPORTED  The driver does not support the analog hardware variant connected.",
	PICO_UNABLE_TO_CONVERT_TO_RESISTANCE:                     "PICO_UNABLE_TO_CONVERT_TO_RESISTANCE  Converting a channel's ADC value to resistance has failed.",
	PICO_DUPLICATED_CHANNEL:                                  "PICO_DUPLICATED_CHANNEL  The channel is listed more than once in the function call.",
	PICO_INVALID_RESISTANCE_CONVERSION:                       "PICO_INVALID_RESISTANCE_CONVERSION  The range cannot have resistance conversion applied.",
	PICO_INVALID_VALUE_IN_MAX_BUFFER:                         "PICO_INVALID_VALUE_IN_MAX_BUFFER  An invalid value is in the max buffer.",
	PICO_INVALID_VALUE_IN_MIN_BUFFER:                         "PICO_INVALID_VALUE_IN_MIN_BUFFER  An invalid value is in the min buffer.",
	PICO_SIGGEN_FREQUENCY_OUT_OF_RANGE:                       "PICO_SIGGEN_FREQUENCY_OUT_OF_RANGE  When calculating the frequency for phase conversion, the frequency is greater than that supported by the current variant.",
	PICO_EEPROM2_CORRUPT:                                     "PICO_EEPROM2_CORRUPT  The device's EEPROM is corrupt. Contact Pico Technology support: https://www.picotech.com/tech-support.",
	PICO_EEPROM2_FAIL:                                        "PICO_EEPROM2_FAIL  The EEPROM has failed.",
	PICO_SERIAL_BUFFER_TOO_SMALL:                             "PICO_SERIAL_BUFFER_TOO_SMALL  The serial buffer is too small for the required information.",
	PICO_SIGGEN_TRIGGER_AND_EXTERNAL_CLOCK_CLASH:             "PICO_SIGGEN_TRIGGER_AND_EXTERNAL_CLOCK_CLASH  The signal generator trigger and the external clock have both been set. This is not allowed.",
	PICO_WARNING_SIGGEN_AUXIO_TRIGGER_DISABLED:               "PICO_WARNING_SIGGEN_AUXIO_TRIGGER_DISABLED  The AUX trigger was enabled and the external clock has been enabled, so the AUX has been automatically disabled.",
	PICO_SIGGEN_GATING_AUXIO_NOT_AVAILABLE:                   "PICO_SIGGEN_GATING_AUXIO_NOT_AVAILABLE  The AUX I/O was set as a scope trigger and is now being set as a signal generator gating trigger. This is not allowed.",
	PICO_SIGGEN_GATING_AUXIO_ENABLED:                         "PICO_SIGGEN_GATING_AUXIO_ENABLED  The AUX I/O was set by the signal generator as a gating trigger and is now being set as a scope trigger. This is not allowed.",
	PICO_RESOURCE_ERROR:                                      "PICO_RESOURCE_ERROR  A resource has failed to initialise.",
	PICO_TEMPERATURE_TYPE_INVALID:                            "PICO_TEMPERATURE_TYPE_INVALID  The temperature type is out of range.",
	PICO_TEMPERATURE_TYPE_NOT_SUPPORTED:                      "PICO_TEMPERATURE_TYPE_NOT_SUPPORTED  A requested temperature type is not supported on this device.",
	PICO_TIMEOUT:                                             "PICO_TIMEOUT  A read/write to the device has timed out.",
	PICO_DEVICE_NOT_FUNCTIONING:                              "PICO_DEVICE_NOT_FUNCTIONING  The device cannot be connected correctly.",
	PICO_INTERNAL_ERROR:                                      "PICO_INTERNAL_ERROR  The driver has experienced an unknown error and is unable to recover from this error.",
	PICO_MULTIPLE_DEVICES_FOUND:                              "PICO_MULTIPLE_DEVICES_FOUND  Used when opening units via IP and more than multiple units have the same IP address.",
	PICO_WARNING_NUMBER_OF_SEGMENTS_REDUCED:                  "PICO_WARNING_NUMBER_OF_SEGMENTS_REDUCED",
	PICO_CAL_PINS_STATES:                                     "PICO_CAL_PINS_STATES  The calibration pin states argument is out of range.",
	PICO_CAL_PINS_FREQUENCY:                                  "PICO_CAL_PINS_FREQUENCY  The calibration pin frequency argument is out of range.",
	PICO_CAL_PINS_AMPLITUDE:                                  "PICO_CAL_PINS_AMPLITUDE  The calibration pin amplitude argument is out of range.",
	PICO_CAL_PINS_WAVETYPE:                                   "PICO_CAL_PINS_WAVETYPE  The calibration pin wavetype argument is out of range.",
	PICO_CAL_PINS_OFFSET:                                     "PICO_CAL_PINS_OFFSET  The calibration pin offset argument is out of range.",
	PICO_PROBE_FAULT:                                         "PICO_PROBE_FAULT  The probe's identity has a problem.",
	PICO_PROBE_IDENTITY_UNKNOWN:                              "PICO_PROBE_IDENTITY_UNKNOWN  The probe has not been identified.",
	PICO_PROBE_POWER_DC_POWER_SUPPLY_REQUIRED:                "PICO_PROBE_POWER_DC_POWER_SUPPLY_REQUIRED  Enabling the probe would cause the device to exceed the allowable current limit.",
	PICO_PROBE_NOT_POWERED_WITH_DC_POWER_SUPPLY:              "PICO_PROBE_NOT_POWERED_WITH_DC_POWER_SUPPLY  The DC power supply is connected; enabling the probe would cause the device to exceed the allowable current limit.",
	PICO_PROBE_CONFIG_FAILURE:                                "PICO_PROBE_CONFIG_FAILURE  Failed to complete probe configuration.",
	PICO_PROBE_INTERACTION_CALLBACK:                          "PICO_PROBE_INTERACTION_CALLBACK  Failed to set the callback function, as currently in current callback function.",
	PICO_UNKNOWN_INTELLIGENT_PROBE:                           "PICO_UNKNOWN_INTELLIGENT_PROBE  The probe has been verified but not known on this driver.",
	PICO_INTELLIGENT_PROBE_CORRUPT:                           "PICO_INTELLIGENT_PROBE_CORRUPT  The intelligent probe cannot be verified.",
	PICO_PROBE_COLLECTION_NOT_STARTED:                        "PICO_PROBE_COLLECTION_NOT_STARTED  The callback is null, probe collection will only start when first callback is a none null pointer.",
	PICO_PROBE_POWER_CONSUMPTION_EXCEEDED:                    "PICO_PROBE_POWER_CONSUMPTION_EXCEEDED  The current drawn by the probe(s) has exceeded the allowed limit.",
	PICO_WARNING_PROBE_CHANNEL_OUT_OF_SYNC:                   "PICO_WARNING_PROBE_CHANNEL_OUT_OF_SYNC  The channel range limits have changed due to connecting or disconnecting a probe the channel has been enabled.",
	PICO_ENDPOINT_MISSING:                                    "PICO_ENDPOINT_MISSING",
	PICO_UNKNOWN_ENDPOINT_REQUEST:                            "PICO_UNKNOWN_ENDPOINT_REQUEST",
	PICO_ADC_TYPE_ERROR:                                      "PICO_ADC_TYPE_ERROR  The ADC on board the device has not been correctly identified.",
	PICO_FPGA2_FAILED:                                        "PICO_FPGA2_FAILED",
	PICO_FPGA2_DEVICE_STATUS:                                 "PICO_FPGA2_DEVICE_STATUS",
	PICO_ENABLE_PROGRAM_FPGA2_FAILED:                         "PICO_ENABLE_PROGRAM_FPGA2_FAILED",
	PICO_NO_CHANNELS_OR_PORTS_ENABLED:                        "PICO_NO_CHANNELS_OR_PORTS_ENABLED",
	PICO_INVALID_RATIO_MODE:                                  "PICO_INVALID_RATIO_MODE",
	PICO_READS_NOT_SUPPORTED_IN_CURRENT_CAPTURE_MODE:         "PICO_READS_NOT_SUPPORTED_IN_CURRENT_CAPTURE_MODE",
	PICO_TRIGGER_READ_SELECTION_CHECK_FAILED:                 "PICO_TRIGGER_READ_SELECTION_CHECK_FAILED  These selection tests can be masked together to show that mode than one read selection has failed the tests, therefore theses error codes cover 0x00000161UL to 0x0000016FUL.",
	PICO_DATA_READ1_SELECTION_CHECK_FAILED:                   "PICO_DATA_READ1_SELECTION_CHECK_FAILED",
	PICO_DATA_READ2_SELECTION_CHECK_FAILED:                   "PICO_DATA_READ2_SELECTION_CHECK_FAILED",
	PICO_DATA_READ3_SELECTION_CHECK_FAILED:                   "PICO_DATA_READ3_SELECTION_CHECK_FAILED",
	PICO_READ_SELECTION_OUT_OF_RANGE:                         "PICO_READ_SELECTION_OUT_OF_RANGE  The requested read is not one of the reads available in enPicoReadSelection.",
	PICO_MULTIPLE_RATIO_MODES:                                "PICO_MULTIPLE_RATIO_MODES  The downsample ratio options cannot be combined together for this request.",
	PICO_NO_SAMPLES_READ:                                     "PICO_NO_SAMPLES_READ  The enPicoReadSelection request has no samples available.",
	PICO_RATIO_MODE_NOT_REQUESTED:                            "PICO_RATIO_MODE_NOT_REQUESTED  The enPicoReadSelection did not include one of the downsample ratios now requested.",
	PICO_NO_USER_READ_REQUESTS_SET:                           "PICO_NO_USER_READ_REQUESTS_SET  No read requests have been made.",
	PICO_ZERO_SAMPLES_INVALID:                                "PICO_ZERO_SAMPLES_INVALID  The parameter for <number of values> cannot be zero.",
	PICO_ANALOGUE_HARDWARE_MISSING:                           "PICO_ANALOGUE_HARDWARE_MISSING  The analog hardware cannot be identified; contact Pico Technology Technical Support.",
	PICO_ANALOGUE_HARDWARE_PINS:                              "PICO_ANALOGUE_HARDWARE_PINS  Setting of the analog hardware pins failed.",
	PICO_ANALOGUE_HARDWARE_SMPS_FAULT:                        "PICO_ANALOGUE_HARDWARE_SMPS_FAULT  An SMPS fault has occurred.",
	PICO_DIGITAL_ANALOGUE_HARDWARE_CONFLICT:                  "PICO_DIGITAL_ANALOGUE_HARDWARE_CONFLICT  There appears to be a conflict between the expected and actual hardware in the device; contact Pico Technology Technical Support.",
	PICO_RATIO_MODE_BUFFER_NOT_SET:                           "PICO_RATIO_MODE_BUFFER_NOT_SET  One or more of the enPicoRatioMode requested do not have a data buffer set.",
	PICO_RESOLUTION_NOT_SUPPORTED_BY_VARIANT:                 "PICO_RESOLUTION_NOT_SUPPORTED_BY_VARIANT  The resolution is valid but not supported by the opened device.",
	PICO_THRESHOLD_OUT_OF_RANGE:                              "PICO_THRESHOLD_OUT_OF_RANGE  The requested trigger threshold is out of range for the current device resolution.",
	PICO_INVALID_SIMPLE_TRIGGER_DIRECTION:                    "PICO_INVALID_SIMPLE_TRIGGER_DIRECTION  The simple trigger only supports upper edge direction options.",
	PICO_AUX_NOT_SUPPORTED:                                   "PICO_AUX_NOT_SUPPORTED  The aux trigger is not supported on this variant.",
	PICO_NULL_DIRECTIONS:                                     "PICO_NULL_DIRECTIONS  The trigger directions pointer may not be null.",
	PICO_NULL_CHANNEL_PROPERTIES:                             "PICO_NULL_CHANNEL_PROPERTIES  The trigger channel properties pointer may not be null.",
	PICO_TRIGGER_CHANNEL_NOT_ENABLED:                         "PICO_TRIGGER_CHANNEL_NOT_ENABLED  A trigger is set on a channel that has not been enabled.",
	PICO_CONDITION_HAS_NO_TRIGGER_PROPERTY:                   "PICO_CONDITION_HAS_NO_TRIGGER_PROPERTY  A trigger condition has been set but a trigger property not set.",
	PICO_RATIO_MODE_TRIGGER_MASKING_INVALID:                  "PICO_RATIO_MODE_TRIGGER_MASKING_INVALID  When requesting trigger data, this option can only be combined with the segment header ratio mode flag.",
	PICO_TRIGGER_DATA_REQUIRES_MIN_BUFFER_SIZE_OF_40_SAMPLES: "PICO_TRIGGER_DATA_REQUIRES_MIN_BUFFER_SIZE_OF_40_SAMPLES  The trigger data buffer must be 40 or more samples in size.",
	PICO_NO_OF_CAPTURES_OUT_OF_RANGE:                         "PICO_NO_OF_CAPTURES_OUT_OF_RANGE  The number of requested waveforms is greater than the number of memory segments allocated.",
	PICO_RATIO_MODE_SEGMENT_HEADER_DOES_NOT_REQUIRE_BUFFERS:  "PICO_RATIO_MODE_SEGMENT_HEADER_DOES_NOT_REQUIRE_BUFFERS  When requesting segment header information, the segment header does not require a data buffer, to get the segment information use GetTriggerInfo.",
	PICO_FOR_SEGMENT_HEADER_USE_GETTRIGGERINFO:               "PICO_FOR_SEGMENT_HEADER_USE_GETTRIGGERINFO  Use GetTriggerInfo to retrieve the segment header information.",
	PICO_READ_NOT_SET:                                        "PICO_READ_NOT_SET  A read request has not been set.",
	PICO_ADC_SETTING_MISMATCH:                                "PICO_ADC_SETTING_MISMATCH  The expected and actual states of the ADCs do not match.",
	PICO_DATATYPE_INVALID:                                    "PICO_DATATYPE_INVALID  The requested data type is not one of the enPicoDataType listed.",
	PICO_RATIO_MODE_DOES_NOT_SUPPORT_DATATYPE:                "PICO_RATIO_MODE_DOES_NOT_SUPPORT_DATATYPE  The down sample ratio mode requested does not support the enPicoDataType option chosen.",
	PICO_CHANNEL_COMBINATION_NOT_VALID_IN_THIS_RESOLUTION:    "PICO_CHANNEL_COMBINATION_NOT_VALID_IN_THIS_RESOLUTION  The channel combination is not valid for the resolution.",
	PICO_USE_8BIT_RESOLUTION:                                 "PICO_USE_8BIT_RESOLUTION",
	PICO_AGGREGATE_BUFFERS_SAME_POINTER:                      "PICO_AGGREGATE_BUFFERS_SAME_POINTER  The buffer for minimum data values and maximum data values are the same buffers.",
	PICO_OVERLAPPED_READ_VALUES_OUT_OF_RANGE:                 "PICO_OVERLAPPED_READ_VALUES_OUT_OF_RANGE  The read request number of samples requested for an overlapped operation are more than the total number of samples to capture.",
	PICO_OVERLAPPED_READ_SEGMENTS_OUT_OF_RANGE:               "PICO_OVERLAPPED_READ_SEGMENTS_OUT_OF_RANGE  The overlapped read request has more segments specified than segments allocated.",
	PICO_CHANNELFLAGSCOMBINATIONS_ARRAY_SIZE_TOO_SMALL:       "PICO_CHANNELFLAGSCOMBINATIONS_ARRAY_SIZE_TOO_SMALL  The number of channel combinations available are greater than the array size received.",
	PICO_CAPTURES_EXCEEDS_NO_OF_SUPPORTED_SEGMENTS:           "PICO_CAPTURES_EXCEEDS_NO_OF_SUPPORTED_SEGMENTS  The number of captures is larger than the maximum number of segments allowed for the device variant.",
	PICO_TIME_UNITS_OUT_OF_RANGE:                             "PICO_TIME_UNITS_OUT_OF_RANGE  The time unit requested is not one of the listed enPicoTimeUnits.",
	PICO_NO_SAMPLES_REQUESTED:                                "PICO_NO_SAMPLES_REQUESTED  The number of samples parameter may not be zero.",
	PICO_INVALID_ACTION:                                      "PICO_INVALID_ACTION  The action requested is not listed in enPicoAction.",
	PICO_NO_OF_SAMPLES_NEED_TO_BE_EQUAL_WHEN_ADDING_BUFFERS:  "PICO_NO_OF_SAMPLES_NEED_TO_BE_EQUAL_WHEN_ADDING_BUFFERS  When adding buffers for the same read request the buffers for all ratio mode requests have to be the same size.",
	PICO_WAITING_FOR_DATA_BUFFERS:                            "PICO_WAITING_FOR_DATA_BUFFERS  The data is being processed but there is no empty data buffers available, a new data buffer needs to be set sent to the driver so that the data can be processed.",
	PICO_STREAMING_ONLY_SUPPORTS_ONE_READ:                    "PICO_STREAMING_ONLY_SUPPORTS_ONE_READ  when streaming data, only one read option is available.",
	PICO_CLEAR_DATA_BUFFER_INVALID:                           "PICO_CLEAR_DATA_BUFFER_INVALID  A clear read request is not one of the enPicoAction listed.",
	PICO_INVALID_ACTION_FLAGS_COMBINATION:                    "PICO_INVALID_ACTION_FLAGS_COMBINATION  The combination of action flags are not allowed.",
	PICO_BOTH_MIN_AND_MAX_NULL_BUFFERS_CANNOT_BE_ADDED:       "PICO_BOTH_MIN_AND_MAX_NULL_BUFFERS_CANNOT_BE_ADDED  PICO_ADD request has been made but both data buffers are set to null and so there is nowhere to put the data.",
	PICO_CONFLICT_IN_SET_DATA_BUFFERS_CALL_REMOVE_DATA_BUFFER_TO_RESET:           "PICO_CONFLICT_IN_SET_DATA_BUFFERS_CALL_REMOVE_DATA_BUFFER_TO_RESET  A conflict between the data buffers being set has occurred. Please use the PICO_CLEAR_ALL action to reset.",
	PICO_REMOVING_DATA_BUFFER_ENTRIES_NOT_ALLOWED_WHILE_DATA_PROCESSING:          "PICO_REMOVING_DATA_BUFFER_ENTRIES_NOT_ALLOWED_WHILE_DATA_PROCESSING  While processing data, buffers cannot be removed from the data buffers list.",
	PICO_TOO_MANY_FREQUENCY_COUNTERS:                                             "PICO_TOO_MANY_FREQUENCY_COUNTERS  An already enabled frequency counter must be disabled before another can be enabled",
	PICO_CYUSB_REQUEST_FAILED:                                                    "PICO_CYUSB_REQUEST_FAILED  An USB request has failed.",
	PICO_STREAMING_DATA_REQUIRED:                                                 "PICO_STREAMING_DATA_REQUIRED  A request has been made to retrieve the latest streaming data, but with either a null pointer or an array size set to zero.",
	PICO_INVALID_NUMBER_OF_SAMPLES:                                               "PICO_INVALID_NUMBER_OF_SAMPLES  A buffer being set has a length that is invalid (ie less than zero).",
	PICO_INVALID_DISTRIBUTION:                                                    "PICO_INVALID_DISTRIBUTION  The distribution size may not be zero.",
	PICO_BUFFER_LENGTH_GREATER_THAN_INT32_T:                                      "PICO_BUFFER_LENGTH_GREATER_THAN_INT32_T  The buffer length in bytes is greater than a 4-byte word.",
	PICO_PLL_MUX_OUT_FAILED:                                                      "PICO_PLL_MUX_OUT_FAILED  The PLL has failed.",
	PICO_ONE_PULSE_WIDTH_DIRECTION_ALLOWED:                                       "PICO_ONE_PULSE_WIDTH_DIRECTION_ALLOWED  Pulse width only supports one direction.",
	PICO_EXTERNAL_TRIGGER_NOT_SUPPORTED:                                          "PICO_EXTERNAL_TRIGGER_NOT_SUPPORTED  There is no external trigger available on the device specified by the handle.",
	PICO_NO_TRIGGER_CONDITIONS_SET:                                               "PICO_NO_TRIGGER_CONDITIONS_SET  The condition parameter is a null pointer.",
	PICO_NO_OF_CHANNEL_TRIGGER_PROPERTIES_OUT_OF_RANGE:                           "PICO_NO_OF_CHANNEL_TRIGGER_PROPERTIES_OUT_OF_RANGE  The number of trigger channel properties it outside the allowed range (is less than zero).",
	PICO_PROBE_COMPONENT_ERROR:                                                   "PICO_PROBE_COMPONENT_ERROR  A probe has been plugged into a channel, but can not be identified correctly.",
	PICO_INCOMPATIBLE_PROBE:                                                      "PICO_INCOMPATIBLE_PROBE  The probe is incompatible with the device channel it is connected to. This could lead to error in the measurements.",
	PICO_INVALID_TRIGGER_CHANNEL_FOR_ETS:                                         "PICO_INVALID_TRIGGER_CHANNEL_FOR_ETS  The requested channel for ETS triggering is not supported.",
	PICO_NOT_AVAILABLE_WHEN_STREAMING_IS_RUNNING:                                 "PICO_NOT_AVAILABLE_WHEN_STREAMING_IS_RUNNING  While the device is streaming the get values method is not available",
	PICO_INVALID_TRIGGER_WITHIN_PRE_TRIGGER_STATE:                                "PICO_INVALID_TRIGGER_WITHIN_PRE_TRIGGER_STATE  the requested state is not one of the enSharedTriggerWithinPreTrigger values",
	PICO_ZERO_NUMBER_OF_CAPTURES_INVALID:                                         "PICO_ZERO_NUMBER_OF_CAPTURES_INVALID  the number of captures have to be greater than zero",
	PICO_INVALID_LENGTH:                                                          "PICO_INVALID_LENGTH  the quantifier for a pointer, defining the length in bytes is invalid",
	PICO_TRIGGER_DELAY_OUT_OF_RANGE:                                              "PICO_TRIGGER_DELAY_OUT_OF_RANGE  the trigger delay is greater than supported by the hardware",
	PICO_INVALID_THRESHOLD_DIRECTION:                                             "PICO_INVALID_THRESHOLD_DIRECTION  the requested threshold direction is not allowed with the specified channel",
	PICO_INVALID_THRESHOLD_MODE:                                                  "PICO_INVALID_THRESHOLD_MODE  the requested threshold mode is not allowed with the specified channel",
	PICO_TIMEBASE_NOT_SUPPORTED_BY_RESOLUTION:                                    "PICO_TIMEBASE_NOT_SUPPORTED_BY_RESOLUTION  The timebase is not supported or is invalid.",
	PICO_INVALID_VARIANT:                                                         "PICO_INVALID_VARIANT  The device variant is not supported by this current driver.",
	PICO_MEMORY_MODULE_ERROR:                                                     "PICO_MEMORY_MODULE_ERROR  The actual memory module does not match the expected memory module.",
	PICO_PULSE_WIDTH_QUALIFIER_LOWER_UPPER_CONFILCT:                              "PICO_PULSE_WIDTH_QUALIFIER_LOWER_UPPER_CONFILCT  A null pointer has been passed in the trigger function or one of the parameters is out of range.",
	PICO_PULSE_WIDTH_QUALIFIER_TYPE:                                              "PICO_PULSE_WIDTH_QUALIFIER_TYPE  The pulse width qualifier type is not one of the listed options.",
	PICO_PULSE_WIDTH_QUALIFIER_DIRECTION:                                         "PICO_PULSE_WIDTH_QUALIFIER_DIRECTION  The pulse width qualifier direction is not one of the listed options.",
	PICO_THRESHOLD_MODE_OUT_OF_RANGE:                                             "PICO_THRESHOLD_MODE_OUT_OF_RANGE  The threshold range is not one of the listed options.",
	PICO_TRIGGER_AND_PULSEWIDTH_DIRECTION_IN_CONFLICT:                            "PICO_TRIGGER_AND_PULSEWIDTH_DIRECTION_IN_CONFLICT  The trigger direction and pulse width option conflict with each other.",
	PICO_THRESHOLD_UPPER_LOWER_MISMATCH:                                          "PICO_THRESHOLD_UPPER_LOWER_MISMATCH  The thresholds upper limits and thresholds lower limits conflict with each other.",
	PICO_PULSE_WIDTH_LOWER_OUT_OF_RANGE:                                          "PICO_PULSE_WIDTH_LOWER_OUT_OF_RANGE  The pulse width lower count is out of range.",
	PICO_PULSE_WIDTH_UPPER_OUT_OF_RANGE:                                          "PICO_PULSE_WIDTH_UPPER_OUT_OF_RANGE  The pulse width upper count is out of range.",
	PICO_FRONT_PANEL_ERROR:                                                       "PICO_FRONT_PANEL_ERROR  The devices front panel has caused an error.",
	PICO_FRONT_PANEL_MODE:                                                        "PICO_FRONT_PANEL_MODE  The actual and expected mode of the front panel do not match.",
	PICO_FRONT_PANEL_FEATURE:                                                     "PICO_FRONT_PANEL_FEATURE  A front panel feature is not available or failed to configure.",
	PICO_NO_PULSE_WIDTH_CONDITIONS_SET:                                           "PICO_NO_PULSE_WIDTH_CONDITIONS_SET  When setting the pulse width conditions either the pointer is null or the number of conditions is set to zero.",
	PICO_TRIGGER_PORT_NOT_ENABLED:                                                "PICO_TRIGGER_PORT_NOT_ENABLED  a trigger condition exists for a port, but the port has not been enabled",
	PICO_DIGITAL_DIRECTION_NOT_SET:                                               "PICO_DIGITAL_DIRECTION_NOT_SET  a trigger condition exists for a port, but no digital channel directions have been set",
	PICO_I2C_DEVICE_INVALID_READ_COMMAND:                                         "PICO_I2C_DEVICE_INVALID_READ_COMMAND",
	PICO_I2C_DEVICE_INVALID_RESPONSE:                                             "PICO_I2C_DEVICE_INVALID_RESPONSE",
	PICO_I2C_DEVICE_INVALID_WRITE_COMMAND:                                        "PICO_I2C_DEVICE_INVALID_WRITE_COMMAND",
	PICO_I2C_DEVICE_ARGUMENT_OUT_OF_RANGE:                                        "PICO_I2C_DEVICE_ARGUMENT_OUT_OF_RANGE",
	PICO_I2C_DEVICE_MODE:                                                         "PICO_I2C_DEVICE_MODE  The actual and expected mode do not match.",
	PICO_I2C_DEVICE_SETUP_FAILED:                                                 "PICO_I2C_DEVICE_SETUP_FAILED  While trying to configure the device, set up failed.",
	PICO_I2C_DEVICE_FEATURE:                                                      "PICO_I2C_DEVICE_FEATURE  A feature is not available or failed to configure.",
	PICO_I2C_DEVICE_VALIDATION_FAILED:                                            "PICO_I2C_DEVICE_VALIDATION_FAILED  The device did not pass the validation checks.",
	PICO_INTERNAL_HEADER_ERROR:                                                   "PICO_INTERNAL_HEADER_ERROR",
	PICO_FAILED_TO_WRITE_HARDWARE_FAULT:                                          "PICO_FAILED_TO_WRITE_HARDWARE_FAULT  The device couldn't write the channel settings due to a hardware fault",
	PICO_MSO_TOO_MANY_EDGE_TRANSITIONS_WHEN_USING_PULSE_WIDTH:                    "PICO_MSO_TOO_MANY_EDGE_TRANSITIONS_WHEN_USING_PULSE_WIDTH  The number of MSO's edge transitions being set is not supported by this device (RISING, FALLING, or RISING_OR_FALLING).",
	PICO_INVALID_PROBE_LED_POSITION:                                              "PICO_INVALID_PROBE_LED_POSITION  A probe LED position requested is not one of the available probe positions in the ProbeLedPosition enum.",
	PICO_PROBE_LED_POSITION_NOT_SUPPORTED:                                        "PICO_PROBE_LED_POSITION_NOT_SUPPORTED  The LED position is not supported by the selected variant.",
	PICO_DUPLICATE_PROBE_CHANNEL_LED_POSITION:                                    "PICO_DUPLICATE_PROBE_CHANNEL_LED_POSITION  A channel has more than one of the same LED position in the ProbeChannelLedSetting struct.",
	PICO_PROBE_LED_FAILURE:                                                       "PICO_PROBE_LED_FAILURE  Setting the probes LED has failed.",
	PICO_PROBE_NOT_SUPPORTED_BY_THIS_DEVICE:                                      "PICO_PROBE_NOT_SUPPORTED_BY_THIS_DEVICE  Probe is not supported by the selected variant.",
	PICO_INVALID_PROBE_NAME:                                                      "PICO_INVALID_PROBE_NAME  The probe name is not in the list of enPicoConnectProbe enums.",
	PICO_NO_PROBE_COLOUR_SETTINGS:                                                "PICO_NO_PROBE_COLOUR_SETTINGS  The number of colour settings are zero or a null pointer passed to the function.",
	PICO_NO_PROBE_CONNECTED_ON_REQUESTED_CHANNEL:                                 "PICO_NO_PROBE_CONNECTED_ON_REQUESTED_CHANNEL  Channel has no probe connected to it.",
	PICO_PROBE_DOES_NOT_REQUIRE_CALIBRATION:                                      "PICO_PROBE_DOES_NOT_REQUIRE_CALIBRATION  Connected probe does not require calibration.",
	PICO_PROBE_CALIBRATION_FAILED:                                                "PICO_PROBE_CALIBRATION_FAILED  Connected probe could not be calibrated - hardware fault is a possible cause.",
	PICO_PROBE_VERSION_ERROR:                                                     "PICO_PROBE_VERSION_ERROR  A probe has been connected, but the version is not recognised.",
	PICO_PROBE_DOES_NOT_SUPPORT_FREQUENCY_COUNTER:                                "PICO_PROBE_DOES_NOT_SUPPORT_FREQUENCY_COUNTER  The channel with the frequency counter enabled has a probe connected that does not support this feature",
	PICO_AUTO_TRIGGER_TIME_TOO_LONG:                                              "PICO_AUTO_TRIGGER_TIME_TOO_LONG  The requested trigger time is to long for the selected variant.",
	PICO_MSO_POD_VALIDATION_FAILED:                                               "PICO_MSO_POD_VALIDATION_FAILED  The MSO pod did not pass the validation checks.",
	PICO_NO_MSO_POD_CONNECTED:                                                    "PICO_NO_MSO_POD_CONNECTED  No MSO pod found on the requested digital port.",
	PICO_DIGITAL_PORT_HYSTERESIS_OUT_OF_RANGE:                                    "PICO_DIGITAL_PORT_HYSTERESIS_OUT_OF_RANGE  the digital port enum value is not in the enPicoDigitalPortHysteresis declaration",
	PICO_MSO_POD_FAILED_UNIT:                                                     "PICO_MSO_POD_FAILED_UNIT",
	PICO_ATTENUATION_FAILED:                                                      "PICO_ATTENUATION_FAILED  The device's EEPROM is corrupt. Contact Pico Technology support: https://www.picotech.com/tech-support.",
	PICO_DC_50OHM_OVERVOLTAGE_TRIPPED:                                            "PICO_DC_50OHM_OVERVOLTAGE_TRIPPED  a channel set to the 50Ohm Path has Tripped due to the input signal",
	PICO_MSO_OVER_CURRENT_TRIPPED:                                                "PICO_MSO_OVER_CURRENT_TRIPPED  The MSO pod over current protection activated, unplug and replug the MSO pod",
	PICO_NOT_RESPONDING_OVERHEATED:                                               "PICO_NOT_RESPONDING_OVERHEATED  Status error for when the device has overheated.",
	PICO_USB_VERSION_NOT_SUPPORTED:                                               "PICO_USB_VERSION_NOT_SUPPORTED  The USB version of the port is not supported by this variant",
	PICO_HARDWARE_CAPTURE_TIMEOUT:                                                "PICO_HARDWARE_CAPTURE_TIMEOUT  waiting for the device to capture timed out",
	PICO_HARDWARE_READY_TIMEOUT:                                                  "PICO_HARDWARE_READY_TIMEOUT  waiting for the device be ready for capture timed out",
	PICO_HARDWARE_CAPTURING_CALL_STOP:                                            "PICO_HARDWARE_CAPTURING_CALL_STOP  the driver is performing a capture requested by RunStreaming or RunBlock to interrupt this capture call Stop on the device first",
	PICO_TOO_FEW_REQUESTED_STREAMING_SAMPLES:                                     "PICO_TOO_FEW_REQUESTED_STREAMING_SAMPLES  the number of samples is less than the minimum number allowed",
	PICO_STREAMING_REREAD_DATA_NOT_AVAILABLE:                                     "PICO_STREAMING_REREAD_DATA_NOT_AVAILABLE  a streaming capture has been made but re-reading the data is not allowed",
	PICO_STREAMING_COMBINATION_OF_RAW_DATA_AND_ONE_AGGREGATION_DATA_TYPE_ALLOWED: "PICO_STREAMING_COMBINATION_OF_RAW_DATA_AND_ONE_AGGREGATION_DATA_TYPE_ALLOWED  When requesting data only Raw and one of the following aggregation data types allowed - PICO_RATIO_MODE_AGGREGATE (Min Max), PICO_RATIO_MODE_DECIMATE, PICO_RATIO_MODE_AVERAGE and/or PICO_RATIO_MODE_SUM, PICO_RATIO_MODE_DISTRIBUTION average and sum are classed as one aggregation type",
	PICO_DEVICE_TIME_STAMP_RESET:                                                 "PICO_DEVICE_TIME_STAMP_RESET  The time stamp per waveform segment has been reset.",
	PICO_TRIGGER_TIME_NOT_REQUESTED:                                              "PICO_TRIGGER_TIME_NOT_REQUESTED  When requesting the TriggerTimeOffset the trigger time has not been set.",
	PICO_TRIGGER_TIME_BUFFER_NOT_SET:                                             "PICO_TRIGGER_TIME_BUFFER_NOT_SET  Trigger time buffer not set.",
	PICO_TRIGGER_TIME_FAILED_TO_CALCULATE:                                        "PICO_TRIGGER_TIME_FAILED_TO_CALCULATE  The trigger time failed to be calculated.",
	PICO_TRIGGER_WITHIN_A_PRE_TRIGGER_FAILED_TO_CALCULATE:                        "PICO_TRIGGER_WITHIN_A_PRE_TRIGGER_FAILED_TO_CALCULATE  The trigger time failed to be calculated.",
	PICO_TRIGGER_TIME_STAMP_NOT_REQUESTED:                                        "PICO_TRIGGER_TIME_STAMP_NOT_REQUESTED  The trigger time stamp was not requested.",
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_REQUIRE_BUFFERS:   "PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_REQUIRE_BUFFERS  RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION cannot have a buffer set",
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_HAVE_BUFFERS:      "PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_DOES_NOT_HAVE_BUFFERS  it is not possible to set a buffer for RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION therefore information is not available pertaining to samples",
	PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_USE_GETTRIGGERINFO:         "PICO_RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION_USE_GETTRIGGERINFO  to get the trigger time use either GetTriggerInfo or GetTriggerTimeOffset api calls",
	PICO_STREAMING_DOES_NOT_SUPPORT_TRIGGER_RATIO_MODES:                          "PICO_STREAMING_DOES_NOT_SUPPORT_TRIGGER_RATIO_MODES  PICO_RATIO_MDOE_TRIGGER and RATIO_MODE_TRIGGER_DATA_FOR_TIME_CALCULATION is not supported in streaming capture",
	PICO_USE_THE_TRIGGER_READ:                                                    "PICO_USE_THE_TRIGGER_READ  only the PICO_TRIGGER_READ may be used to read PICO_RATIO_MODE_TRIGGER, and PICO_RATIO_MODE_TRIGGER_FOR_CALCULATION",
	PICO_USE_A_DATA_READ:                                                         "PICO_USE_A_DATA_READ  one of the PICO_DATA_READs should be used to read: PICO_RATIO_MODE_RAW PICO_RATIO_MODE_AGGREGATE PICO_RATIO_MODE_DECIMATE PICO_RATIO_MODE_AVERAGE",
	PICO_TRIGGER_READ_REQUIRES_INT16_T_DATA_TYPE:                                 "PICO_TRIGGER_READ_REQUIRES_INT16_T_DATA_TYPE  trigger data always requires a PICO_INT16_T data type",
	PICO_RATIO_MODE_REQUIRES_NUMBER_OF_SAMPLES_TO_BE_SET:                         "PICO_RATIO_MODE_REQUIRES_NUMBER_OF_SAMPLES_TO_BE_SET  a ratio mode passed to the API call requires the number of samples to be greater than zero",
	PICO_SIGGEN_SETTINGS_MISMATCH:                                                "PICO_SIGGEN_SETTINGS_MISMATCH  Attempted to set up the signal generator with an inconsistent configuration.",
	PICO_SIGGEN_SETTINGS_CHANGED_CALL_APPLY:                                      "PICO_SIGGEN_SETTINGS_CHANGED_CALL_APPLY  The signal generator has been partially reconfigured and the new settings must be applied before it can be paused or restarted.",
	PICO_SIGGEN_WAVETYPE_NOT_SUPPORTED:                                           "PICO_SIGGEN_WAVETYPE_NOT_SUPPORTED  The wave type is not listed in enPicoWaveType.",
	PICO_SIGGEN_TRIGGERTYPE_NOT_SUPPORTED:                                        "PICO_SIGGEN_TRIGGERTYPE_NOT_SUPPORTED  The trigger type is not listed in enSigGenTrigType.",
	PICO_SIGGEN_TRIGGERSOURCE_NOT_SUPPORTED:                                      "PICO_SIGGEN_TRIGGERSOURCE_NOT_SUPPORTED  The trigger source is not listed in enSigGenTrigSource.",
	PICO_SIGGEN_FILTER_STATE_NOT_SUPPORTED:                                       "PICO_SIGGEN_FILTER_STATE_NOT_SUPPORTED  The filter state is not listed in enPicoSigGenFilterState.",
	PICO_SIGGEN_NULL_PARAMETER:                                                   "PICO_SIGGEN_NULL_PARAMETER  The arbitrary waveform buffer is a null pointer.",
	PICO_SIGGEN_EMPTY_BUFFER_SUPPLIED:                                            "PICO_SIGGEN_EMPTY_BUFFER_SUPPLIED  The arbitrary waveform buffer length is zero.",
	PICO_SIGGEN_RANGE_NOT_SUPPLIED:                                               "PICO_SIGGEN_RANGE_NOT_SUPPLIED  The sig gen voltage offset and peak to peak have not been set.",
	PICO_SIGGEN_BUFFER_NOT_SUPPLIED:                                              "PICO_SIGGEN_BUFFER_NOT_SUPPLIED  The sig gen arbitrary waveform buffer not been set.",
	PICO_SIGGEN_FREQUENCY_NOT_SUPPLIED:                                           "PICO_SIGGEN_FREQUENCY_NOT_SUPPLIED  The sig gen frequency have not been set.",
	PICO_SIGGEN_SWEEP_INFO_NOT_SUPPLIED:                                          "PICO_SIGGEN_SWEEP_INFO_NOT_SUPPLIED  The sig gen sweep information has not been set.",
	PICO_SIGGEN_TRIGGER_INFO_NOT_SUPPLIED:                                        "PICO_SIGGEN_TRIGGER_INFO_NOT_SUPPLIED  The sig gen trigger information has not been set.",
	PICO_SIGGEN_CLOCK_FREQ_NOT_SUPPLIED:                                          "PICO_SIGGEN_CLOCK_FREQ_NOT_SUPPLIED  The sig gen clock frequency have not been set.",
	PICO_SIGGEN_TOO_MANY_SAMPLES:                                                 "PICO_SIGGEN_TOO_MANY_SAMPLES  The sig gen arbitrary waveform buffer length is too long.",
	PICO_SIGGEN_DUTYCYCLE_OUT_OF_RANGE:                                           "PICO_SIGGEN_DUTYCYCLE_OUT_OF_RANGE  The duty cycle value is out of range.",
	PICO_SIGGEN_CYCLES_OUT_OF_RANGE:                                              "PICO_SIGGEN_CYCLES_OUT_OF_RANGE  The number of cycles is out of range.",
	PICO_SIGGEN_PRESCALE_OUT_OF_RANGE:                                            "PICO_SIGGEN_PRESCALE_OUT_OF_RANGE  The pre-scaler is out of range.",
	PICO_SIGGEN_SWEEPTYPE_INVALID:                                                "PICO_SIGGEN_SWEEPTYPE_INVALID  The sweep type is not listed in enPicoSweepType.",
	PICO_SIGGEN_SWEEP_WAVETYPE_MISMATCH:                                          "PICO_SIGGEN_SWEEP_WAVETYPE_MISMATCH  A mismatch has occurred while checking the sweeps wave type.",
	PICO_SIGGEN_INVALID_SWEEP_PARAMETERS:                                         "PICO_SIGGEN_INVALID_SWEEP_PARAMETERS  The sweeps or shots and trigger type are not valid when combined together.",
	PICO_SIGGEN_SWEEP_PRESCALE_NOT_SUPPORTED:                                     "PICO_SIGGEN_SWEEP_PRESCALE_NOT_SUPPORTED  The sweep and prescaler are not valid when combined together.",
	PICO_AWG_OVER_VOLTAGE_RANGE:                                                  "PICO_AWG_OVER_VOLTAGE_RANGE  The potential applied to the AWG output exceeds the maximum voltage range of the AWG.",
	PICO_NOT_LOCKED_TO_REFERENCE_FREQUENCY:                                       "PICO_NOT_LOCKED_TO_REFERENCE_FREQUENCY  The reference signal cannot be locked to.",
	PICO_PERMISSIONS_ERROR:                                                       "PICO_PERMISSIONS_ERROR  (Linux only.) udev rules are incorrectly configured. The user does not have read/write permissions on the device's file descriptor.",
	PICO_PORTS_WITHOUT_ANALOGUE_CHANNELS_ONLY_ALLOWED_IN_8BIT_RESOLUTION:         "PICO_PORTS_WITHOUT_ANALOGUE_CHANNELS_ONLY_ALLOWED_IN_8BIT_RESOLUTION  The digital ports without analog channels are only allowed in 8-bit resolution.",
	PICO_ANALOGUE_FRONTEND_MISSING:                                               "PICO_ANALOGUE_FRONTEND_MISSING",
	PICO_FRONT_PANEL_MISSING:                                                     "PICO_FRONT_PANEL_MISSING",
	PICO_ANALOGUE_FRONTEND_AND_FRONT_PANEL_MISSING:                               "PICO_ANALOGUE_FRONTEND_AND_FRONT_PANEL_MISSING",
	PICO_DIGITAL_BOARD_HARDWARE_ERROR:                                            "PICO_DIGITAL_BOARD_HARDWARE_ERROR  The digital board has reported an error to the driver",
	PICO_FIRMWARE_UPDATE_REQUIRED_TO_USE_DEVICE_WITH_THIS_DRIVER:                 "PICO_FIRMWARE_UPDATE_REQUIRED_TO_USE_DEVICE_WITH_THIS_DRIVER  checking if the firmware needs updating the updateRequired parameter is null",
	PICO_UPDATE_REQUIRED_NULL:                                                    "PICO_UPDATE_REQUIRED_NULL",
	PICO_FIRMWARE_UP_TO_DATE:                                                     "PICO_FIRMWARE_UP_TO_DATE",
	PICO_FLASH_FAIL:                                                              "PICO_FLASH_FAIL",
	PICO_INTERNAL_ERROR_FIRMWARE_LENGTH_INVALID:                                  "PICO_INTERNAL_ERROR_FIRMWARE_LENGTH_INVALID",
	PICO_INTERNAL_ERROR_FIRMWARE_NULL:                                            "PICO_INTERNAL_ERROR_FIRMWARE_NULL",
	PICO_FIRMWARE_FAILED_TO_BE_CHANGED:                                           "PICO_FIRMWARE_FAILED_TO_BE_CHANGED",
	PICO_FIRMWARE_FAILED_TO_RELOAD:                                               "PICO_FIRMWARE_FAILED_TO_RELOAD",
	PICO_FIRMWARE_FAILED_TO_BE_UPDATE:                                            "PICO_FIRMWARE_FAILED_TO_BE_UPDATE",
	PICO_FIRMWARE_VERSION_OUT_OF_RANGE:                                           "PICO_FIRMWARE_VERSION_OUT_OF_RANGE",
	PICO_OPTIONAL_BOOTLOADER_UPDATE_AVAILABLE_WITH_THIS_DRIVER:                   "PICO_OPTIONAL_BOOTLOADER_UPDATE_AVAILABLE_WITH_THIS_DRIVER",
	PICO_BOOTLOADER_VERSION_NOT_AVAILABLE:                                        "PICO_BOOTLOADER_VERSION_NOT_AVAILABLE",
	PICO_NO_APPS_AVAILABLE:                                                       "PICO_NO_APPS_AVAILABLE",
	PICO_UNSUPPORTED_APP:                                                         "PICO_UNSUPPORTED_APP",
	PICO_ADC_POWERED_DOWN:                                                        "PICO_ADC_POWERED_DOWN  the adc is powered down when trying to capture data",
	PICO_WATCHDOGTIMER:                                                           "PICO_WATCHDOGTIMER  An internal error has occurred and a watchdog timer has been called.",
	PICO_IPP_NOT_FOUND:                                                           "PICO_IPP_NOT_FOUND  The picoipp.dll has not been found.",
	PICO_IPP_NO_FUNCTION:                                                         "PICO_IPP_NO_FUNCTION  A function in the picoipp.dll does not exist.",
	PICO_IPP_ERROR:                                                               "PICO_IPP_ERROR  The Pico IPP call has failed.",
	PICO_SHADOW_CAL_NOT_AVAILABLE:                                                "PICO_SHADOW_CAL_NOT_AVAILABLE  Shadow calibration is not available on this device.",
	PICO_SHADOW_CAL_DISABLED:                                                     "PICO_SHADOW_CAL_DISABLED  Shadow calibration is currently disabled.",
	PICO_SHADOW_CAL_ERROR:                                                        "PICO_SHADOW_CAL_ERROR  Shadow calibration error has occurred.",
	PICO_SHADOW_CAL_CORRUPT:                                                      "PICO_SHADOW_CAL_CORRUPT  The shadow calibration is corrupt.",
	PICO_DEVICE_MEMORY_OVERFLOW:                                                  "PICO_DEVICE_MEMORY_OVERFLOW  The memory on board the device has overflowed.",
	PICO_ADC_TEST_FAILURE:                                                        "PICO_ADC_TEST_FAILURE  The device Adc test failed.",
	PICO_RESERVED_1:                                                              "PICO_RESERVED_1",
	PICO_SOURCE_NOT_READY:                                                        "PICO_SOURCE_NOT_READY  The PicoSource device is not ready to accept instructions.",
	PICO_SOURCE_INVALID_BAUD_RATE:                                                "PICO_SOURCE_INVALID_BAUD_RATE",
	PICO_SOURCE_NOT_OPENED_FOR_WRITE:                                             "PICO_SOURCE_NOT_OPENED_FOR_WRITE",
	PICO_SOURCE_FAILED_TO_WRITE_DEVICE:                                           "PICO_SOURCE_FAILED_TO_WRITE_DEVICE",
	PICO_SOURCE_EEPROM_FAIL:                                                      "PICO_SOURCE_EEPROM_FAIL",
	PICO_SOURCE_EEPROM_NOT_PRESENT:                                               "PICO_SOURCE_EEPROM_NOT_PRESENT",
	PICO_SOURCE_EEPROM_NOT_PROGRAMMED:                                            "PICO_SOURCE_EEPROM_NOT_PROGRAMMED",
	PICO_SOURCE_LIST_NOT_READY:                                                   "PICO_SOURCE_LIST_NOT_READY",
	PICO_SOURCE_FTD2XX_NOT_FOUND:                                                 "PICO_SOURCE_FTD2XX_NOT_FOUND",
	PICO_SOURCE_FTD2XX_NO_FUNCTION:                                               "PICO_SOURCE_FTD2XX_NO_FUNCTION",
}

func StatStr(code int) string {
	return statMap[code]
}
