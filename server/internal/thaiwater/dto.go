package thaiwater

// The DTOs below mirror the parts of the JSON the product uses. Fields that
// ADR 0003 forbids reading (left_bank and right_bank) are deliberately not
// declared, so no code can use them by accident. The published
// situation_level, storage_percent and diff_wl_bank are declared only so the
// tests can use them as an oracle.

// Text is a field that ThaiWater publishes in Thai and English.
type Text struct {
	TH string `json:"th"`
	EN string `json:"en"`
}

// LoadResponse is the body of GET .../public/waterlevel_load.
type LoadResponse struct {
	WaterlevelData struct {
		Result string    `json:"result"`
		Data   []LoadRow `json:"data"`
	} `json:"waterlevel_data"`
	Scale struct {
		Result string `json:"result"`
		Data   struct {
			// RuleWeb holds the thresholds the website applies to
			// storage_percent. (scale.data.scale holds labels and colours.)
			RuleWeb []Rule `json:"rule_web"`
		} `json:"data"`
	} `json:"scale"`
}

// Rule is one threshold of scale.data.rule_web, such as {">", "100", 5}.
type Rule struct {
	Operator string `json:"operator"`
	Term     string `json:"term"`
	Level    int    `json:"level"`
}

// LoadRow is one station's latest reading in the national snapshot.
type LoadRow struct {
	ID            int64  `json:"id"`
	Datetime      string `json:"waterlevel_datetime"` // "2006-01-02 15:04", ICT, no offset
	WaterlevelMSL Num    `json:"waterlevel_msl"`      // metres above mean sea level
	// Oracle fields: compared in tests, never used for product values.
	StoragePercent Num    `json:"storage_percent"`
	SituationLevel *int   `json:"situation_level"` // absent on rows without references
	DiffWLBank     Num    `json:"diff_wl_bank"`    // unsigned; see ADR 0003
	DiffWLBankText string `json:"diff_wl_bank_text"`

	Station   Station `json:"station"`
	Geocode   Geocode `json:"geocode"`
	Basin     Basin   `json:"basin"`
	Agency    Agency  `json:"agency"`
	RiverName string  `json:"river_name"` // absent on about one row in ten
}

// Station is the station block inside a load row.
type Station struct {
	ID          int64  `json:"id"`
	Name        Text   `json:"tele_station_name"`
	Lat         Num    `json:"tele_station_lat"`
	Lon         Num    `json:"tele_station_long"`
	OldCode     string `json:"tele_station_oldcode"` // the code people know, such as "C.35"
	MinBank     Num    `json:"min_bank"`             // bank-full (ADR 0003), 2 decimals here
	GroundLevel Num    `json:"ground_level"`         // lower reference; may be absent
}

// Geocode is the administrative location of a station.
type Geocode struct {
	ProvinceCode string `json:"province_code"`
	Province     Text   `json:"province_name"`
	Amphoe       Text   `json:"amphoe_name"`
	Tumbon       Text   `json:"tumbon_name"`
}

// Basin is the river basin a station belongs to.
type Basin struct {
	Name Text `json:"basin_name"`
}

// Agency is the organisation that operates a station.
type Agency struct {
	Short Text `json:"agency_shortname"`
}

// GraphResponse is the body of GET .../public/waterlevel_graph.
type GraphResponse struct {
	Result string `json:"result"`
	Data   struct {
		GraphData []GraphPoint `json:"graph_data"`
		// Full-precision references, preferred over the load's 2-decimal
		// values (ADR 0003). RID values carry float32 noise (40.599998).
		MinBank     Num `json:"min_bank"`
		GroundLevel Num `json:"ground_level"`
	} `json:"data"`
}

// GraphPoint is one history value: every 10 minutes for HII stations,
// hourly for RID stations.
type GraphPoint struct {
	Datetime string `json:"datetime"` // "2006-01-02 15:04", ICT
	Value    Num    `json:"value"`
}
