package ingestion

import "strings"

// Geographic keyword dict
var countryDictionary = map[string]string{
	// TW
	"TAIWAN": "TW",
	"JIUFEN": "TW",

	// US States & Territories (USGS frequently lists states directly)
	"PR":                       "PR",
	"VI":                       "US",
	"GU":                       "GU",
	"AS":                       "AS",
	"MP":                       "MP",
	"ALASKA":                   "US",
	"HAWAII":                   "US",
	"COLORADO":                 "US",
	"CALIFORNIA":               "US",
	"NEW MEXICO":               "US",
	"VIRGIN ISLANDS":           "US",
	"OKLAHOMA":                 "US",
	"NEVADA":                   "US",
	"WASHINGTON":               "US",
	"OREGON":                   "US",
	"IDAHO":                    "US",
	"UTAH":                     "US",
	"ARIZONA":                  "US",
	"MONTANA":                  "US",
	"WYOMING":                  "US",
	"TEXAS":                    "US",
	"KANSAS":                   "US",
	"NEBRASKA":                 "US",
	"SOUTH DAKOTA":             "US",
	"NORTH DAKOTA":             "US",
	"MINNESOTA":                "US",
	"IOWA":                     "US",
	"MISSOURI":                 "US",
	"ARKANSAS":                 "US",
	"LOUISIANA":                "US",
	"WISCONSIN":                "US",
	"ILLINOIS":                 "US",
	"KENTUCKY":                 "US",
	"TENNESSEE":                "US",
	"MISSISSIPPI":              "US",
	"ALABAMA":                  "US",
	"MICHIGAN":                 "US",
	"INDIANA":                  "US",
	"OHIO":                     "US",
	"FLORIDA":                  "US",
	"SOUTH CAROLINA":           "US",
	"NORTH CAROLINA":           "US",
	"VIRGINIA":                 "US",
	"WEST VIRGINIA":            "US",
	"MARYLAND":                 "US",
	"DELAWARE":                 "US",
	"PENNSYLVANIA":             "US",
	"NEW JERSEY":               "US",
	"NEW YORK":                 "US",
	"CONNECTICUT":              "US",
	"RHODE ISLAND":             "US",
	"MASSACHUSETTS":            "US",
	"VERMONT":                  "US",
	"NEW HAMPSHIRE":            "US",
	"MAINE":                    "US",
	"GUAM":                     "GU",
	"AMERICAN SAMOA":           "AS",
	"NORTHERN MARIANA ISLANDS": "MP",

	// Countries & Regions
	"AFGHANISTAN":              "AF",
	"ALBANIA":                  "AL",
	"ALGERIA":                  "DZ",
	"ANDORRA":                  "AD",
	"ANGOLA":                   "AO",
	"ARGENTINA":                "AR",
	"ARMENIA":                  "AM",
	"AUSTRALIA":                "AU",
	"AUSTRIA":                  "AT",
	"AZERBAIJAN":               "AZ",
	"BAHAMAS":                  "BS",
	"BAHRAIN":                  "BH",
	"BANGLADESH":               "BD",
	"BARBADOS":                 "BB",
	"BELARUS":                  "BY",
	"BELGIUM":                  "BE",
	"BELIZE":                   "BZ",
	"BENIN":                    "BJ",
	"BHUTAN":                   "BT",
	"BOLIVIA":                  "BO",
	"BOSNIA AND HERZEGOVINA":   "BA",
	"BOTSWANA":                 "BW",
	"BRAZIL":                   "BR",
	"BRUNEI":                   "BN",
	"BULGARIA":                 "BG",
	"BURKINA FASO":             "BF",
	"BURUNDI":                  "BI",
	"CABO VERDE":               "CV",
	"CAMBODIA":                 "KH",
	"CAMEROON":                 "CM",
	"CANADA":                   "CA",
	"CENTRAL AFRICAN REPUBLIC": "CF",
	"CHAD":                     "TD",
	"CHILE":                    "CL",
	"CHINA":                    "CN",
	"COLOMBIA":                 "CO",
	"COMOROS":                  "KM",
	"CONGO":                    "CG",
	"COSTA RICA":               "CR",
	"CROATIA":                  "HR",
	"CUBA":                     "CU",
	"CYPRUS":                   "CY",
	"CZECH REPUBLIC":           "CZ",
	"DENMARK":                  "DK",
	"DJIBOUTI":                 "DJ",
	"DOMINICA":                 "DM",
	"DOMINICAN REPUBLIC":       "DO",
	"ECUADOR":                  "EC",
	"EGYPT":                    "EG",
	"EL SALVADOR":              "SV",
	"EQUATORIAL GUINEA":        "GQ",
	"ERITREA":                  "ER",
	"ESTONIA":                  "EE",
	"ESWATINI":                 "SZ",
	"ETHIOPIA":                 "ET",
	"FIJI":                     "FJ",
	"FINLAND":                  "FI",
	"FRANCE":                   "FR",
	"GABON":                    "GA",
	"GAMBIA":                   "GM",
	"GEORGIA":                  "GE",
	"GERMANY":                  "DE",
	"GHANA":                    "GH",
	"GREECE":                   "GR",
	"GRENADA":                  "GD",
	"GUATEMALA":                "GT",
	"GUINEA":                   "GN",
	"GUINEA-BISSAU":            "GW",
	"GUYANA":                   "GY",
	"HAITI":                    "HT",
	"HONDURAS":                 "HN",
	"HUNGARY":                  "HU",
	"ICELAND":                  "IS",
	"INDIA":                    "IN",
	"INDONESIA":                "ID",
	"IRAN":                     "IR",
	"IRAQ":                     "IQ",
	"IRELAND":                  "IE",
	"ISRAEL":                   "IL",
	"ITALY":                    "IT",
	"JAMAICA":                  "JM",
	"JAPAN":                    "JP",
	"JORDAN":                   "JO",
	"KAZAKHSTAN":               "KZ",
	"KENYA":                    "KE",
	"KIRIBATI":                 "KI",
	"KOSOVA":                   "XK",
	"KOSOVO":                   "XK",
	"KUWAIT":                   "KW",
	"KYRGYZSTAN":               "KG",
	"LAOS":                     "LA",
	"LATVIA":                   "LV",
	"LEBANON":                  "LB",
	"LESOTHO":                  "LS",
	"LIBERIA":                  "LR",
	"LIBYA":                    "LY",
	"LIECHTENSTEIN":            "LI",
	"LITHUANIA":                "LT",
	"LUXEMBOURG":               "LU",
	"MADAGASCAR":               "MG",
	"MALAWI":                   "MW",
	"MALAYSIA":                 "MY",
	"MALDIVES":                 "MV",
	"MALI":                     "ML",
	"MALTA":                    "MT",
	"MARSHALL ISLANDS":         "MH",
	"MAURITANIA":               "MR",
	"MAURITIUS":                "MU",
	"MEXICO":                   "MX",
	"MICRONESIA":               "FM",
	"MOLDOVA":                  "MD",
	"MONACO":                   "MC",
	"MONGOLIA":                 "MN",
	"MONTENEGRO":               "ME",
	"MOROCCO":                  "MA",
	"MOZAMBIQUE":               "MZ",
	"MYANMAR":                  "MM",
	"NAMIBIA":                  "NA",
	"NAURU":                    "NR",
	"NEPAL":                    "NP",
	"NETHERLANDS":              "NL",
	"NEW ZEALAND":              "NZ",
	"NICARAGUA":                "NI",
	"NIGER":                    "NE",
	"NIGERIA":                  "NG",
	"NORTH MACEDONIA":          "MK",
	"NORWAY":                   "NO",
	"OMAN":                     "OM",
	"PAKISTAN":                 "PK",
	"PALAU":                    "PW",
	"PALESTINE":                "PS",
	"PANAMA":                   "PA",
	"PAPUA NEW GUINEA":         "PG",
	"PARAGUAY":                 "PY",
	"PERU":                     "PE",
	"PHILIPPINES":              "PH",
	"POLAND":                   "PL",
	"PORTUGAL":                 "PT",
	"PUERTO RICO":              "PR",
	"QATAR":                    "QA",
	"ROMANIA":                  "RO",
	"RUSSIA":                   "RU",
	"RWANDA":                   "RW",
	"SAMOA":                    "WS",
	"SAN MARINO":               "SM",
	"SAO TOME AND PRINCIPE":    "ST",
	"SAUDI ARABIA":             "SA",
	"SENEGAL":                  "SN",
	"SERBIA":                   "RS",
	"SEYCHELLES":               "SC",
	"SIERRA LEONE":             "SL",
	"SINGAPORE":                "SG",
	"SLOVAKIA":                 "SK",
	"SLOVENIA":                 "SI",
	"SOLOMON ISLANDS":          "SB",
	"SOMALIA":                  "SO",
	"SOUTH AFRICA":             "ZA",
	"SOUTH KOREA":              "KR",
	"SOUTH SUDAN":              "SS",
	"SPAIN":                    "ES",
	"SRI LANKA":                "LK",
	"SUDAN":                    "SD",
	"SURINAME":                 "SR",
	"SWEDEN":                   "SE",
	"SWITZERLAND":              "CH",
	"SYRIA":                    "SY",
	"TAJIKISTAN":               "TJ",
	"TANZANIA":                 "TZ",
	"THAILAND":                 "TH",
	"TIMOR-LESTE":              "TL",
	"TIMOR LESTE":              "TL",
	"TOGO":                     "TG",
	"TONGA":                    "TO",
	"TRINIDAD AND TOBAGO":      "TT",
	"TUNISIA":                  "TN",
	"TURKEY":                   "TR",
	"TURKIYE":                  "TR",
	"TURKMENISTAN":             "TM",
	"TUVALU":                   "TV",
	"UGANDA":                   "UG",
	"UKRAINE":                  "UA",
	"UNITED ARAB EMIRATES":     "AE",
	"UNITED KINGDOM":           "GB",
	"UNITED STATES":            "US",
	"URUGUAY":                  "UY",
	"UZBEKISTAN":               "UZ",
	"VANUATU":                  "VU",
	"VATICAN CITY":             "VA",
	"VENEZUELA":                "VE",
	"VIETNAM":                  "VN",
	"YEMEN":                    "YE",
	"ZAMBIA":                   "ZM",
	"ZIMBABWE":                 "ZW",
}

var usStates = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true, "CT": true, "DE": true,
	"FL": true, "GA": true, "HI": true, "ID": true, "IL": true, "IN": true, "IA": true, "KS": true,
	"KY": true, "LA": true, "ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
	"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true, "NM": true, "NY": true,
	"NC": true, "ND": true, "OH": true, "OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
	"SD": true, "TN": true, "TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
	"WI": true, "WY": true,
}

// Transfer Country name
func parseCountryFromPlace(place string) string {
	if place == "" {
		return "UNKNOWN"
	}

	placeUpper := strings.ToUpper(place)

	// USGS places are typically structured as: "[distance] [direction] of [location], [Country or US State]"
	// If a comma is present, try to match the trimmed last part exactly first.
	if idx := strings.LastIndex(placeUpper, ","); idx != -1 {
		lastPart := strings.TrimSpace(placeUpper[idx+1:])

		// 1. Direct exact match in our dictionary (e.g. "CANADA" -> "CA", "PR" -> "PR")
		if isoCode, exists := countryDictionary[lastPart]; exists {
			return isoCode
		}

		// 2. Check if it's a 2-letter US State postal abbreviation (e.g. "CA", "AK", "NV")
		if usStates[lastPart] {
			return "US"
		}

		// 3. Direct exact match for USA indicator
		if lastPart == "USA" || lastPart == "UNITED STATES" {
			return "US"
		}
	}

	// Fallback to substring matching if exact match of the last part is not found
	for keyword, isoCode := range countryDictionary {
		if strings.Contains(placeUpper, keyword) {
			return isoCode
		}
	}

	// Suffix/Indicator checks
	if strings.Contains(placeUpper, "USA") || strings.Contains(placeUpper, "UNITED STATES") {
		return "US"
	}

	// Ocean indicators (e.g., Ridge, Trench, Basin, Ocean, Rise)
	if strings.Contains(placeUpper, "RIDGE") ||
		strings.Contains(placeUpper, "TRENCH") ||
		strings.Contains(placeUpper, "BASIN") ||
		strings.Contains(placeUpper, "OCEAN") ||
		strings.Contains(placeUpper, "RISE") {
		return "OCEAN"
	}

	return "OCEAN"
}
