package parsers

import (
	"fmt"
	"strings"
)

// MinecraftLanguage represents a Minecraft language with its code, native name, and English name
type MinecraftLanguage struct {
	Code    string `json:"code"`
	Native  string `json:"native"`
	English string `json:"english"`
}

// MinecraftLanguages contains all supported Minecraft languages
var MinecraftLanguages = []MinecraftLanguage{
	{"af_za", "Afrikaans (Suid-Afrika)", "Afrikaans"},
	{"ar_sa", "العربية (العالم العربي)", "Arabic"},
	{"ast_es", "Asturianu (Asturies)", "Asturian"},
	{"az_az", "Azərbaycanca (Azərbaycan)", "Azerbaijani"},
	{"ba_ru", "Башҡортса (Башҡортостан)", "Bashkir"},
	{"bar", "Boarisch", "Bavarian"},
	{"be_by", "Беларуская (Беларусь)", "Belarusian"},
	{"bg_bg", "Български (България)", "Bulgarian"},
	{"br_fr", "Brezhoneg (Breizh)", "Breton"},
	{"brb", "Barbadian Creole", "Barbadian Creole"},
	{"bs_ba", "Bosanski (Bosna i Hercegovina)", "Bosnian"},
	{"ca_es", "Català (Catalunya)", "Catalan"},
	{"cs_cz", "Čeština (Česká republika)", "Czech"},
	{"cy_gb", "Cymraeg (Cymru)", "Welsh"},
	{"da_dk", "Dansk (Danmark)", "Danish"},
	{"de_at", "Deutsch (Österreich)", "German (Austria)"},
	{"de_ch", "Deutsch (Schweiz)", "German (Switzerland)"},
	{"de_de", "Deutsch (Deutschland)", "German (Germany)"},
	{"el_gr", "Ελληνικά (Ελλάδα)", "Greek"},
	{"en_au", "English (Australia)", "English (Australia)"},
	{"en_ca", "English (Canada)", "English (Canada)"},
	{"en_gb", "English (United Kingdom)", "English (United Kingdom)"},
	{"en_nz", "English (New Zealand)", "English (New Zealand)"},
	{"en_pt", "Pirate Speak", "Pirate Speak"},
	{"en_ud", "ǝƃɐnƃuɐl ɥsᴉlƃuǝ", "English (Upside Down)"},
	{"en_us", "English (US)", "English (US)"},
	{"enp", "ǝnbᴉʇuɐ sᴉɥ┴", "Neapolitan"},
	{"enws", "Early Modern English", "Early Modern English"},
	{"eo_uy", "Esperanto", "Esperanto"},
	{"es_ar", "Español (Argentina)", "Spanish (Argentina)"},
	{"es_cl", "Español (Chile)", "Spanish (Chile)"},
	{"es_ec", "Español (Ecuador)", "Spanish (Ecuador)"},
	{"es_es", "Español (España)", "Spanish (Spain)"},
	{"es_mx", "Español (México)", "Spanish (Mexico)"},
	{"es_uy", "Español (Uruguay)", "Spanish (Uruguay)"},
	{"es_ve", "Español (Venezuela)", "Spanish (Venezuela)"},
	{"esan", "Esánski", "Esan"},
	{"et_ee", "Eesti (Eesti)", "Estonian"},
	{"eu_es", "Euskera (Euskadi)", "Basque"},
	{"fa_ir", "فارسی (ایران)", "Persian"},
	{"fi_fi", "Suomi (Suomi)", "Finnish"},
	{"fil_ph", "Filipino (Pilipinas)", "Filipino"},
	{"fo_fo", "Føroyskt (Føroyar)", "Faroese"},
	{"fr_ca", "Français (Canada)", "French (Canada)"},
	{"fr_fr", "Français (France)", "French (France)"},
	{"fra_de", "Fränkisch (Franken)", "Franconian"},
	{"fy_nl", "Frysk (Fryslân)", "Frisian"},
	{"ga_ie", "Gaeilge (Éire)", "Irish"},
	{"gd_gb", "Gàidhlig (Alba)", "Scottish Gaelic"},
	{"gl_es", "Galego (Galicia)", "Galician"},
	{"got_de", "𐌲𐌿𐍄𐌹𐍃𐌺", "Gothic"},
	{"gv_im", "Gaelg (Ellan Vannin)", "Manx"},
	{"haw_us", "ʻŌlelo Hawaiʻi (Hawaiʻi)", "Hawaiian"},
	{"he_il", "עברית (ישראל)", "Hebrew"},
	{"hi_in", "हिन्दी (भारत)", "Hindi"},
	{"hr_hr", "Hrvatski (Hrvatska)", "Croatian"},
	{"hu_hu", "Magyar (Magyarország)", "Hungarian"},
	{"hy_am", "Հայերեն (Հայաստան)", "Armenian"},
	{"id_id", "Bahasa Indonesia (Indonesia)", "Indonesian"},
	{"ig_ng", "Igbo (Nigeria)", "Igbo"},
	{"io_en", "Ido", "Ido"},
	{"is_is", "Íslenska (Ísland)", "Icelandic"},
	{"isv", "Medžuslovjansky", "Interslavic"},
	{"it_it", "Italiano (Italia)", "Italian"},
	{"ja_jp", "日本語 (日本)", "Japanese"},
	{"jbo_en", "la .lojban.", "Lojban"},
	{"ka_ge", "ქართული (საქართველო)", "Georgian"},
	{"kk_kz", "Қазақша (Қазақстан)", "Kazakh"},
	{"kn_in", "ಕನ್ನಡ (ಭಾರತ)", "Kannada"},
	{"ko_kr", "한국어 (대한민국)", "Korean"},
	{"ksh", "Kölsch", "Ripuarian"},
	{"kw_gb", "Kernowek (Kernow)", "Cornish"},
	{"la_la", "Latina", "Latin"},
	{"lb_lu", "Lëtzebuergesch (Lëtzebuerg)", "Luxembourgish"},
	{"li_li", "Limburgs", "Limburgish"},
	{"lmo", "Lombard", "Lombard"},
	{"lol_us", "LOLCAT", "LOLCAT"},
	{"lt_lt", "Lietuvių (Lietuva)", "Lithuanian"},
	{"lv_lv", "Latviešu (Latvija)", "Latvian"},
	{"lzh", "文言文", "Literary Chinese"},
	{"mk_mk", "Македонски (Македонија)", "Macedonian"},
	{"mn_mn", "Монгол (Монгол)", "Mongolian"},
	{"ms_my", "Bahasa Melayu (Malaysia)", "Malay"},
	{"mt_mt", "Malti (Malta)", "Maltese"},
	{"nds_de", "Plattdüütsch (Düütschland)", "Low German"},
	{"nl_be", "Nederlands (België)", "Dutch (Belgium)"},
	{"nl_nl", "Nederlands (Nederland)", "Dutch (Netherlands)"},
	{"nn_no", "Norsk nynorsk (Noreg)", "Norwegian Nynorsk"},
	{"no_no", "Norsk bokmål (Norge)", "Norwegian Bokmål"},
	{"oc_fr", "Occitan (França)", "Occitan"},
	{"ovd", "Övdalsk", "Elfdalian"},
	{"pl_pl", "Polski (Polska)", "Polish"},
	{"pt_br", "Português (Brasil)", "Portuguese (Brazil)"},
	{"pt_pt", "Português (Portugal)", "Portuguese (Portugal)"},
	{"qya_aa", "Quenya", "Quenya"},
	{"ro_ro", "Română (România)", "Romanian"},
	{"rpr", "Kitrall'", "Kitrall"},
	{"ru_ru", "Русский (Россия)", "Russian"},
	{"ry_ua", "Русиньскый (Русиньско)", "Rusyn"},
	{"sah_sah", "Саха тыла (Саха сирэ)", "Sakha"},
	{"se_no", "Davvisámegiella (Norga)", "Northern Sami"},
	{"sk_sk", "Slovenčina (Slovensko)", "Slovak"},
	{"sl_si", "Slovenščina (Slovenija)", "Slovenian"},
	{"so_so", "Soomaali (Soomaaliya)", "Somali"},
	{"sq_al", "Shqip (Shqipëria)", "Albanian"},
	{"sr_sp", "Српски (Србија)", "Serbian"},
	{"sv_se", "Svenska (Sverige)", "Swedish"},
	{"swg", "Schwäbisch", "Swabian"},
	{"sxu", "Schläsisch", "Upper Saxon"},
	{"szl", "Ślōnskŏ gŏdka", "Silesian"},
	{"ta_in", "தமிழ் (இந்தியா)", "Tamil"},
	{"th_th", "ไทย (ประเทศไทย)", "Thai"},
	{"tl_ph", "Tagalog (Pilipinas)", "Tagalog"},
	{"tlh_aa", "tlhIngan Hol", "Klingon"},
	{"tok", "toki pona", "Toki Pona"},
	{"tr_tr", "Türkçe (Türkiye)", "Turkish"},
	{"tt_ru", "Татарча (Россия)", "Tatar"},
	{"uk_ua", "Українська (Україна)", "Ukrainian"},
	{"val_es", "Valencià (País Valencià)", "Valencian"},
	{"vec_it", "Vèneto", "Venetian"},
	{"vi_vn", "Tiếng Việt (Việt Nam)", "Vietnamese"},
	{"yi_de", "ײִדיש (דײַטשלאַנד)", "Yiddish"},
	{"yo_ng", "Yorùbá (Nàìjíríà)", "Yoruba"},
	{"zh_cn", "简体中文（中国大陆）", "Chinese (Simplified)"},
	{"zh_hk", "繁體中文（香港特別行政區）", "Chinese (Hong Kong)"},
	{"zh_tw", "繁體中文（台灣）", "Chinese (Traditional)"},
	{"zlm_arab", "بهاس ملايو", "Malay (Arabic script)"},
}

// languageMap maps language codes to their data for quick lookup
var languageMap map[string]MinecraftLanguage

func init() {
	languageMap = make(map[string]MinecraftLanguage)
	for _, lang := range MinecraftLanguages {
		languageMap[lang.Code] = lang
	}
}

// ValidateLanguageCode checks if a given language code is supported by Minecraft
func ValidateLanguageCode(code string) bool {
	_, exists := languageMap[strings.ToLower(code)]
	return exists
}

// GetLanguage returns the language data for a given code
func GetLanguage(code string) (MinecraftLanguage, bool) {
	lang, exists := languageMap[strings.ToLower(code)]
	return lang, exists
}

// GetSupportedLanguageCodes returns a slice of all supported language codes
func GetSupportedLanguageCodes() []string {
	codes := make([]string, len(MinecraftLanguages))
	for i, lang := range MinecraftLanguages {
		codes[i] = lang.Code
	}
	return codes
}

// FormatLanguageCodeForVersion formats a language code according to Minecraft version
func FormatLanguageCodeForVersion(code, minecraftVersion string) (string, error) {
	if !ValidateLanguageCode(code) {
		return "", fmt.Errorf("unsupported language code: %s", code)
	}

	// Normalize to lowercase first
	normalizedCode := strings.ToLower(code)

	// For Minecraft 1.10.2 and earlier, use mixed case format
	if isLegacyVersion(minecraftVersion) {
		parts := strings.Split(normalizedCode, "_")
		if len(parts) == 2 {
			return fmt.Sprintf("%s_%s", strings.ToLower(parts[0]), strings.ToUpper(parts[1])), nil
		}
	}

	// For Minecraft 1.11+, use lowercase format
	return normalizedCode, nil
}

// GetLanguageNameForPrompt returns a human-readable language name for translation prompts
func GetLanguageNameForPrompt(code string) string {
	if lang, exists := GetLanguage(code); exists {
		// Prefer English name for clarity in prompts, fallback to native name
		if lang.English != "" {
			return lang.English
		}
		return lang.Native
	}
	return code // fallback to code if not found
}

// isLegacyVersion checks if the Minecraft version is 1.10.2 or earlier
func isLegacyVersion(version string) bool {
	// Simple version comparison for major versions
	legacyVersions := []string{"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9", "1.10"}
	for _, legacy := range legacyVersions {
		if strings.HasPrefix(version, legacy) {
			return true
		}
	}
	return false
}
